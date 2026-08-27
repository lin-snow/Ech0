# 配置与设置的依赖架构（config / setting / kvstore / transaction）

> 本文讲清 Ech0 里「配置」相关的四层——`config`（env + 默认值）、`internal/setting`（设置原语 + seeder）、`kvstore`（KV 存储）、`transaction`（事务）——**谁依赖谁、事务怎么流、读写各走哪条路**。这是一道严肃的架构题：分层划错、读写不分，会直接拖垮可维护性。
>
> 配套：分层后端总览见 [architecture-overview.md](architecture-overview.md)；事务/DI 约定见 `CLAUDE.md`。

---

## 1. 为什么有「两层配置」，且不能合并

Ech0 的配置分两层，刻意不合并，因为它们的**生命周期**和**可变性**根本不同：

| | **Bootstrap 配置** | **运行时设置** |
|---|---|---|
| 存在哪 | env / `internal/config/config.go` | KV 设置表（经 `durableKV`） |
| 典型项 | DB 路径/类型、端口、Host、data root、**JWT 密钥**、日志 | 站点标题、是否开放注册、S3、OAuth2、Agent、Embedding、Comment… |
| 何时被读 | **进程启动早期**，在 DB / 设置表就绪*之前* | 运行期，任意请求 |
| 谁能改 | 部署者（改 env 后重启） | 管理员（后台「系统设置」页，热生效） |
| 能否搬进设置页 | **不能** | **应该** |

**为什么 bootstrap 配置搬不进设置页** —— 鸡生蛋：设置表本身就在 DB 里。要打开 DB 你得先知道 DB 路径；要在 DB 没起来前返回错误页你得先有端口。这些参数在「能读设置表」之前就被用到了，只能来自 env / config。

> **JWT 密钥是这层最容易踩的坑。** 它不走 caarlos0/env 的 tag，而是裸 `os.Getenv("JWT_SECRET")`（`config.go:320` 的 `getJWTSecret`）。**不设就每次启动随机生成**——结果是每次重启所有已签发 token 全部失效。所以生产环境务必固化 `JWT_SECRET`。这也是「为什么少数项必须留在 env」最直白的例子。

`config.go` 用 `caarlos0/env` 给约 84 个字段挂了 `env:"ECH0_*"` tag，技术上端口、S3、日志、上传上限、事件运行时参数都能用 env 覆盖。**但「能覆盖」不等于「推荐这么用」**：除了上面的 bootstrap 项，其余用户可调项都应走设置页（落 KV），env 只是兜底。`config.Config()` 是 `sync.Once` 单例，全局只解析一次。

---

## 2. 分层依赖全景

依赖**只向下**，无环。每一层只认它下面那层的接口：

```
            ┌─────────────────────────────────────────────────┐
  最底层     │ env ─→ config.go（解析 env + 兜底默认值，单例）    │  无依赖，被人依赖
  （地基）   │ model（纯结构体）        gorm + SQLite（外部驱动） │
            └─────────────────────────────────────────────────┘
                     │ Spec.Default() 读取             │
                     ▼                                 ▼
            ┌────────────────────────┐      ┌──────────────────────────┐
  基础设施   │ internal/setting        │      │ transaction.Transactor    │
            │  Spec[T] + Get/Set/Seed │      │  Run(ctx,fn): 开事务，把   │
            │  事务透明：只转发 ctx    │      │  *gorm.DB 塞进 ctx(TxKey) │
            └────────────────────────┘      └──────────────────────────┘
                     │ 调 kv.Get/Set(ctx)              │ tx 随 ctx 向下流
                     ▼                                 ▼
            ┌────────────────────────┐      ┌──────────────────────────┐
            │ kvstore.Store(durableKV)│─────▶│ KeyValueRepository.getDB: │
            │  Persistent 委托 Backend │ 委托  │  ctx 有 tx 就入伙该事务，  │
            └────────────────────────┘      │  没有就用普通连接(自动提交)│
                                            │  读路径带读穿透缓存，事务  │
                                            │  内自动绕过(见 §3)          │
                                            └──────────────────────────┘
```

各层职责与边界：

- **`config`**：唯一的「默认值之源」。`internal/setting` 每个设置项的 `Spec.Default()` 直接读 `config.Config()`（`registry.go:28/72/95…`）。方向**恒为 `config → value`，绝不反向写回 config**——KV 没值时回退到 config 默认，有值则以 KV 为准。
- **`transaction.Transactor`**：唯一**开启**事务的地方。`GormTransactor.Run` 用 GORM 自动事务，把事务态的 `*gorm.DB` 放进 `ctx`（`TxKey`）。事务靠 **`context.Context` 传播**，下游谁都不用「持有」它。
- **`internal/setting`**：设置的**裸原语层**（详见 §3）。import 只向下——`setting.go` 只 import `kvstore`；`registry.go` 另加 `config/i18n/model/util`。**绝不 import `transaction` / service / handler**。
- **`kvstore`**：统一 KV 抽象。`kvstore.Persistent` 委托给 Backend（`KeyValueRepository`），把底层 `gorm.ErrRecordNotFound` 归一化成 `kvstore.ErrNotFound`，对上屏蔽持久化细节。
- **repo（`KeyValueRepository`）**：真正落库的地方，也是事务的**汇合点**——`getDB(ctx)` 从 `ctx` 捞事务（`keyvalue.go:32`）。repo **只用事务，从不开事务**（呼应「事务在 service 层」约定）。

---

## 3. `internal/setting` 为什么是「事务透明的裸原语」

它是基础设施，不是 service、也不是 repo。它对事务的态度是**透明转发**而非「管理」：

```go
// internal/setting/setting.go
func Get[T any](ctx context.Context, kv kvstore.Store, spec Spec[T]) (T, error) { ... }
func Set[T any](ctx context.Context, kv kvstore.Store, spec Spec[T], value T) error { ... }
```

- 它收到啥 `ctx` 就往 `kv.Get/Set(ctx, …)` 传啥。
- 给它一个「事务里的 ctx」→ 写自动并入该事务；给它普通 ctx → 各自提交。
- **开不开事务由调用方（service）通过「传不传带 tx 的 ctx」决定**，不是它的职责。

这也是为什么 `Get` **永不需要事务**（纯读、自带归一化、缺失回退 `Default()`、无副作用、幂等），而**写**的事务需求由调用方按场景决定：

- 单次 `Set` 是一次 upsert，单语句本身就原子，**不强制要事务**；
- 真正需要事务的是 **read-modify-write 复合**或**一次写多个 key**——而那应该发生在 service 层（实测分布与判断口诀见 §4「进阶：写设置时要不要 transactor」）。

配套细节：`KeyValueRepository` 的读路径带读穿透缓存，但 `ReadThroughTypedUnlessTx` 在 `transaction.HasTx(ctx)` 为真时**直接走 DB 绕过缓存**（`cache/patterns.go:25`），避免事务内读到事务外的旧缓存。这让「事务随 ctx 流下去」这条链在缓存层也保持一致。

**Seed 不归 service。** `setting.Seed(ctx, durableKV)` 由应用生命周期层在 `BeforeStart` 直接调（`internal/app/provider.go:33`），启动期把缺失的 key 幂等落库一次（绝不覆盖用户值），**绕开 SettingService**。所以 SettingService 只用到 `internal/setting` 的 **Get / Set**，不碰 Seed。

---

## 4. 核心规约：**读直连，写走域**

这是整套设计落到日常编码的一条规矩，代码里也是这么执行的：

```
coreSetting.Set  →  只出现在 internal/service/setting/*（全部 9 处都在 setting 域内）
coreSetting.Get  →  散落各处：storage/manager、task/scheduled、auth、comment、
                    embedding、connect、event/subscriber/agent、setting 域 …
```

**为什么读能直连：** 两个理由，第二个是硬约束：

1. `Get` 是纯读、无副作用、无原子性顾虑，任意层持有 `durableKV` 即可直接调，不必为读去拉 `SettingService`。
2. **断 DI 构造环。** 跨域读设置时，若依赖整个 `SettingService`，很容易撞上构造期循环——`SettingService` 用到某域、该域又回头依赖 `SettingService`。直接 `coreSetting.Get(ctx, durableKV, …)` 从根上切断它。`internal/task/scheduled/snapshot.go:26` 的注释就是现成例子：「计划配置统一经 setting 引擎读 durableKV（而非依赖整个 SettingService），从根上断开『SettingService → Snapshot → SettingService』的构造环」。所以「读直连」常常不是风格选择，而是**唯一能避免 Wire 成环的写法**。

**为什么写要走 `SettingService`：** 「改设置」几乎从不是「单纯写一下」。看 `SettingService` 持有的依赖就懂了——它握着 `storageManager` / `webhookSender` / `tokenRevoker` / `bus`。一次设置写通常裹着：

1. **鉴权** —— 谁有权改（admin 校验）；
2. **原子性** —— read-modify-write 要包在一个事务里；
3. **副作用扇出** —— 重配 S3 存储管理器、吊销 token、发 webhook、清缓存。

这些全是 service 层的活，`coreSetting.Set` 一概不管（它的注释原话：「鉴权/校验/副作用属调用方职责，不在此原语内」）。把写集中在 setting 域 = **让设置表只有一个 writer**，副作用不会四散漏写。

### 决策表：我要读/写设置，该调谁？

| 我要做的事 | 怎么做 | 依赖什么 |
|---|---|---|
| 读某个设置块（如系统设置、Comment、Embedding 配置） | 直接 `coreSetting.Get(ctx, durableKV, coreSetting.Xxx)` | 持有 `durableKV kvstore.Store` |
| 改设置（带鉴权/原子性/副作用） | 注入并调用 `SettingService` 的对应方法 | 依赖 `setting` 域服务 |
| 启动期补齐默认值 | 不用自己做，`setting.Seed` 已在 `BeforeStart` 处理 | —— |
| 真·零鉴权零副作用的单 key 裸写（极罕见） | 技术上可 `coreSetting.Set`，但请先确认确实无任何副作用 | —— |

### 进阶：写设置时要不要 `transactor`？

有个常见误解：「用 `Get` 安全，用 `Set` 就得引入 transactor」。**不对——触发 transactor 的不是「调了 Set」，而是「先读再改写」。** `Get` 永远无需事务；`Set` 要不要事务，取决于这次写是哪一种：

| 写法 | 例子 | 要 `transactor`？ | 原因 |
|---|---|---|---|
| **整块覆盖**：前端提交完整对象 → 用 DTO 直接构造一个全新的 `model.XxxSetting` → 一次 `Set` | `UpdateAgentSettings`、`UpdateEmbeddingSetting`、oauth2/passkey/snapshot 的 Update | **否** | 单语句 upsert 本身就原子，无「读」可被并发穿插 |
| **read-modify-write**：先 `Get` 现值 → 改其中一部分 / 按条件 → 写回；或要和别的写一起原子 | `UpdateS3Setting`（读改写 + 重配存储管理器）、`BootstrapDefaultLocale`（条件式读改写） | **是** | 读与写之间若被并发写穿插会丢更新，必须裹进同一个 `transactor.Run` |

判断口诀:**「我是整块覆盖，还是先读再改写 / 要和别的写一起原子？」** 只有后者才 `transactor.Run`。注意这个判断发生在 **`SettingService` 内部**——按 §4 规约，setting 域外的代码本就不该直接调 `coreSetting.Set`，自然也不会在域外纠结 transactor。鉴权（admin 校验）所有 Update 都做，但那与事务正交，别混为一谈。

### 谁才真的依赖 `SettingService`？（避免下次重新排查）

承上：既然读直连、写走域，那「依赖 `SettingService`（领域服务）」的就应该很少。全仓核对，真正**注入并调用其方法**的只有 3 处，且没有一处是「为了纯读」：

| 依赖方 | 用它做什么 | 性质 |
|---|---|---|
| `internal/handler/setting` | `GetSetting`/`UpdateSetting`/`GetS3Setting`… | **域内**，就是正常的 handler→service 主轴 |
| `internal/mcp`（adapter） | webhook 的 `GetAll`/`Create`/`Update`/`Delete`/`Test`（`adapter_webhook.go`） | **跨域**：把 webhook 管理暴露给外部 LLM，读+写 |
| `internal/service/init`（InitService） | `BootstrapDefaultLocale`（写：首次建站把部署者语言设为站点默认） | **跨域**：首次建站编排器，见 §5 |

> **import 了 `internal/service/setting` ≠ 依赖 `SettingService`。** 还有几处文件 import 它，但都不是行为依赖，别误判：
> - `internal/repository/setting`、`internal/repository/webhook`：`var _ settingService.SettingRepository = (*…)(nil)` / `WebhookRepository`——这是 setting 服务在 `ports.go` 声明的「我需要的仓储长这样」**端口接口**，repo 在编译期断言自己满足它。箭头朝内（依赖倒置），与「依赖 SettingService」方向相反。
> - `internal/repository/provider.go`、`internal/service/provider.go`、`internal/di/wire_gen.go`：纯 Wire 接线（`wire.Bind` / `NewSettingService`）。
>
> 其余领域（storage/manager、task/scheduled/snapshot、embedding、connect、auth、comment、user、event/subscriber/agent）一律 `coreSetting.Get` 直连，**没有一个为读去依赖 `SettingService`**——所以本文这套规约是对现状的如实描述，不是待办。

---

## 5. 案例：让 UserService「只依赖 internal/setting」时，那次设置写该安在哪

目标：让 `user` 域只经 `coreSetting.Get` 直读、**不依赖 `SettingService`**。`UserService` 原本用 `SettingService` 两件事——`Register` 读 `AllowRegister`（纯读，直接换 `coreSetting.Get` 即可），以及 `InitOwner` 调 `BootstrapDefaultLocale`（写）。难点全在后者：它是 setting 域的**写行为**，按 §4「写走域」不能内联进 user。

`BootstrapDefaultLocale`（`internal/service/setting/system_setting_service.go:35`）是一段 **read-modify-write 原子操作**：

```go
return s.transactor.Run(ctx, func(ctx context.Context) error {        // ← 开事务
    current, err := coreSetting.Get(ctx, s.durableKV, coreSetting.System)
    if err != nil { return err }
    if i18nUtil.ResolveLocale(current.DefaultLocale) != string(commonModel.DefaultLocale) {
        return nil                                                     // ← 站长已手动改过就不覆盖
    }
    current.DefaultLocale = resolved
    return coreSetting.Set(ctx, s.durableKV, coreSetting.System, current)
})
```

考虑过三条路，前两条都被否：

- **❌ 把 RMW 内联进 UserService（直接用 `coreSetting`）**：`coreSetting` 没有 `BootstrapDefaultLocale`，只有 `Get`/`Set`。你得把「开事务 + 判断不覆盖 + 写回」整段**抄进 user 域**——重复逻辑，且把 setting 域的「站点默认语言怎么落库」策略**泄漏**进用户域，违反「写走域」。
- **❌ 做成 `UserCreated` 事件订阅者**：订阅者要调 `BootstrapDefaultLocale`，就得在**精简的事件注入器**（`BuildEventRegistrar`）里构建整个 `SettingService`——它依赖 `*storage.Manager`（顶层共享单例，得穿参进来）+ auth/webhook/file/common 一堆子图，把本该精简的注入器搞臃肿；否则只能让订阅者自己 `coreSetting.Set`，又破坏「单 writer」。
- **✅ 把「调用」上移到 `InitService`**：写仍留在 `SettingService.BootstrapDefaultLocale`（写走域不变），只是**谁来调它**从 `UserService.InitOwner` 上移到 `InitService.InitOwner`。`InitService` 在 `HandlerSet` 里构建，`SettingService` 现成——**零 DI 膨胀、同步、不破坏分层**。

落地结果：

- `UserService`：删掉 `BootstrapDefaultLocale` 调用；`Register` 改 `coreSetting.Get`；字段 `settingService` 换成 `durableKV`。**自此只依赖 `internal/setting`。**
- `InitService`：注入 `SettingService`，在 `InitOwner` 里 `userService.InitOwner(...)` 成功后调 `BootstrapDefaultLocale`（best-effort，失败仅告警）。它本就是「首次建站编排器」，协调 user（建 owner）+ setting（站点默认语言）两域，这是该行为的天然归属。

**心法：** 当某域想摆脱对 `SettingService` 的依赖时，先分清它要的是**读**还是**写行为**。读直接 `coreSetting.Get`；写不能内联（违反写走域），而该把**调用点**安放到一个「本就持有 `SettingService` 的编排层」——而不是硬把 `SettingService` 拖进一个不该有它的精简注入器。`InitService` 之于建站，正是这样的编排层。

---

## 速查

- 找默认值 → `internal/config/config.go` + `internal/setting/registry.go` 的 `Spec.Default()`。
- 加一个新设置项 → 在 `registry.go` 加 `Spec[T]`（key + Default + Normalize），seeder 自动落库；读处直接 `coreSetting.Get`，写处走 `SettingService`。
- 事务为什么「凭空生效」→ `transactor.Run` 把 tx 塞进 `ctx`，repo 的 `getDB(ctx)` 捞出来用；中间各层只转发 `ctx`。
- token 重启就失效 → `JWT_SECRET` 没设，被 `getJWTSecret` 随机生成了（`config.go:320`）。
