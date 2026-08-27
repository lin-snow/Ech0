# Ech0 Job Runner（长时有状态作业子系统）设计

> 状态：**草案 / 主要决策已收敛**（命名、数据模型、接口、进度策略、A–E 已定，见各节）。
> 关联文档：`docs/dev/logging.md`（module 字段约定）、`docs/dev/timezone-design.md`（时间戳为 UTC epoch）。
> 涉及现有代码：`internal/service/migrator`（迁移）、`internal/service/embedding` + `internal/handler/embedding`（重建索引）、`internal/task`（Tasker）。

## 1. 背景与动机

当前仓库里「异步干活」散在三种互不相干的形态里，其中**第 2 类**（长时、有状态、可观察、可取消的一次性作业）缺少统一载体：

| # | 概念 | 形态 | 现状载体 |
|---|------|------|----------|
| 1 | 定时任务 (Scheduled Task) | timer/cron 触发、周期、每次无持久状态、fire-and-forget | `internal/task.Tasker` + gocron |
| 2 | **作业 (Job)** | 一次性、按需触发、长时运行、生命周期可观察（pending→running→success/failed/cancelled）、可取消、状态可查 | migration 用 `GlobalMigrationStateDTO` 存 KeyValue；另有一张**废弃的 `MigrationJob` 表** |
| 3 | 事件驱动异步 | publish/subscribe（尽力投递，无持久重试） | Busen |

第 2 类原本**只有 migration 一个真实客户**，实现是「单全局状态塞 KeyValue + 一次性 goroutine + 前端轮询」。现在出现**第二个真实客户**：

- **reindex（重建向量索引）** 当前是**同步阻塞 HTTP**：`EmbeddingHandler.Reindex()` 直接 `embeddingService.Backfill(ctx)` 死等整条 page 循环跑完才返回。历史 Echo 一多必然超时。它需要改成「后台异步 + 前端轮询进度」——**与 migration 形态完全一致**。

两个真实客户到位，抽象有了依据（而非过早设计）。本文设计 `internal/job` 子系统，把第 2 类统一起来。

### 1.1 核心洞见：触发器 vs 工作单元

「定时任务 vs 作业」并非对立的两类，而是**触发器（何时干）vs 工作单元（干什么）**——前者调度后者。Tasker 不该自己内联业务逻辑，它只是众多触发器之一：

```
            ┌──────────── Triggers（何时）────────────┐
  定时 (Tasker/gocron) ──┐
  用户点击 (HTTP API)    ──┼──► jobService.Submit("reindex", payload)
  事件 (Busen)          ──┘
```

一个「每周自动重建索引」的定时任务，从此就是一行 `jobService.Submit("reindex", …)`，而不是一坨内联逻辑。这就是「定时任务和 job 解耦」的落地。

## 2. 目标与非目标

### 2.1 目标

- **G1**：提供 `internal/job` 子系统，统一承载长时有状态作业的**状态机 / goroutine 生命周期 / 取消 / 持久化 / 状态查询**，各作业只实现「怎么干活」。
- **G2**：把 **reindex 改成异步 job**，前端复用 migration 那套轮询范式展示进度。
- **G3**：把 **migration 迁到同一子系统**，删除其手写状态机（`saveGlobalStateWithRetry`/`runGlobalMigration`/`activeCancel`），消除「两套状态机并存」。
- **G4**：**触发器解耦**——HTTP / Tasker / Busen 都通过 `jobService.Submit` 提交作业，Tasker 不再内联业务逻辑。
- **G5**：保持轻量——**一张通用 `jobs` 表（每 type 单行）、零新增重依赖、本期不暴露作业历史/列表 API/UI**。

### 2.2 非目标（本期明确搁置）

- **作业历史**：每 type 仅保留当前/最近一次状态（单行 upsert，见 §5.2）。**不留多轮历史**，不做 list 端点与历史 UI。
- **分布式 / 多实例调度**：Ech0 是单进程自部署，作业活在内存 + 单库，不考虑多实例抢占。
- **作业重试 / 持久队列**：失败即失败，由用户重新触发；带重试的异步走第 3 类（Busen），不在本子系统。
- **并发同类型作业**：每 type 同时只允许一条非终态作业（§7.3 的 DB 约束直接保证）。并发留作后续。
- **进度跨重启续显**：fine-grained 进度只活内存，崩溃即弃（§8）。这是有意为之，不是缺陷。
- **backup-export 迁入**：潜在第三客户，本期不动；设计须保证它只是「再加一个 Runner」（§14-E）。

## 3. 设计原则

1. **触发器与工作单元分离**：`Service` 只管生命周期，`Runner` 只管干活，触发器只管 `Submit`。三者互不知道彼此细节。
2. **领域 payload 不透明**：通用 `Job` 只存生命周期字段；领域专属的输入/进度/结果序列化进 JSON，`Service` **绝不解析它**，只有对应 `Runner` 和前端认得。
3. **泛型在边缘、擦除在中心**：异构注册表强制边界 untyped，但每个 Runner 通过 `Adapt[P]` 对着 typed 结构体编写（§6.2）。
4. **加法而非替换**：迁移 migration 时**保持其对前端的 API 契约不变**（尤其 `idle` 哨兵，§9.2），轮询响应形状只增不破。
5. **不静默吞错**：作业失败必须落 `Status=failed` + `Error` 原因，沿用 Backfill 现有「全军覆没回传底层错误」的实践。
6. **持久价值决定是否落库**：只有 durable 的生命周期转换落 DB；无持久价值的 fine-grained 进度只进内存（§8）。

## 4. 架构总览

```
┌──────────────────────────── Triggers ────────────────────────────┐
│  HTTP: EmbeddingHandler.Reindex / MigratorHandler.StartMigration  │
│  Tasker (gocron): 未来「定时重建/定时迁移」                          │
│  Busen subscriber: 未来事件触发                                     │
└───────────────────────────────┬───────────────────────────────────┘
                                 │ Submit(type, payload) / Get(type) / Cancel(type)
┌──────────────────────── internal/job ─────────────────────────────┐
│  job.Service                                                      │
│   · runners map[type]Runner    （注册表）                          │
│   · repo    JobRepository      （durable：jobs 表，每 type 单行）  │
│   · live    map[type]*Progress （内存：实时进度，不落库）            │
│   · cancels map[type]CancelFunc（在跑作业的取消句柄）               │
│   Submit: 互斥校验 → upsert(pending) → 同步登记 cancel → go run     │
│   Get:    durable 行 ⊕ 叠加内存实时进度                            │
│   Cancel: 触发 ctx 取消，Runner 协作退出                           │
└───────────────┬───────────────────────────────┬───────────────────┘
                │ Run(ctx, payload, report)       │ durable 转换落库
        ┌───────┴────────┐                ┌──────┴───────┐
   ReindexRunner      MigrationRunner      jobs 表 (GORM)
   (embedding.Backfill   (migrator ETL       PK=type / status /
    拆成可上报 phase)      管道)               phase / error /
                                             payload / started/finished
                                 ▲
                                 │ GET .../status  (前端按 type 轮询)
┌──────────────────────────── Frontend ─────────────────────────────┐
│  复用 migration 既有轮询范式：拿 status + 解析 payload 渲染进度条     │
└────────────────────────────────────────────────────────────────────┘
```

## 5. 数据模型

### 5.1 删除 `MigrationJob` 死表

`internal/model/migration/job.go` 的 `MigrationJob` + `internal/repository/migration` 的 `CreateJob/UpdateJob/GetJobByID` **零调用者**（已核实），是误导性死代码——让人误以为「有个 job 在轮询」，实则运行态走 KeyValue。且它带 `SourceType` 这种迁移专属字段、无通用 `Type`，**无法原样复活成通用表**。

处置：
- 删 `MigrationJob` model、删 repo 的 `CreateJob/UpdateJob/GetJobByID` 三个死方法、从 `internal/database/database.go:165` 的 `AutoMigrate(...)` 列表移除 `&migrationModel.MigrationJob{}`。
- 迁移相关常量（`MigrationStatusXxx`/`MigrationPhaseXxx`）按需保留或下沉到 `MigrationRunner`（§9）。
- DB 层：`migration_jobs` 表既已进过 `AutoMigrate`，老库里有残留空表。GORM 不自动 DROP，**倾向留着不管**（无害、零风险），不补 drop 迁移。

### 5.2 通用 `Job` 表（`internal/job` 拥有，每 type 单行）

```go
// internal/job/model.go
type Job struct {
    Type       string `gorm:"primaryKey;size:64" json:"type"`         // 主键：每 type 仅一行  ★
    Status     string `gorm:"type:varchar(32);index" json:"status"`   // pending/running/success/failed/cancelled
    Phase      string `gorm:"type:varchar(64)" json:"phase"`          // 终态时的阶段快照，可空
    Error      string `gorm:"type:text" json:"error"`
    Payload    string `gorm:"type:text" json:"payload"`               // 领域 JSON blob，Service 不解析
    StartedAt  *int64 `json:"started_at"`
    FinishedAt *int64 `json:"finished_at"`
    UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at"`
}
```

- **主键即 `Type`**：结构性地保证「每 type 单行」，新一次 `Submit` 直接 upsert 覆盖旧终态行。无 id、无历史，最省、最好管。
- 加入 `internal/database/database.go` 的 `AutoMigrate` 列表。

**为什么用表而非沿用 KeyValue 单槽**（重新讲诚实——此处不再以「历史/id 身份」为由，那两条因单行设计已不成立）：

> 表的价值是 **「所有 job 类型共用一套 typed schema + 一个 `JobRepository` + 一个 `job.Service`」**——reindex / migration / 未来 backup 不必各自发明 KeyValue 键 + 专属 DTO。`Type` 主键的唯一性还让 §7.3 的「同类型互斥」由 DB 约束白送（upsert 即原子覆盖；非终态判断在事务内做）。这条理由独立于行数与历史，单行设计下依旧成立。

> 代价：新一次 run 开跑即覆盖上一次 run 的终态结果——与 migration 现状（重开迁移、旧态即弃）完全一致，可接受。

### 5.3 `Payload` 列承载两个客户（验证无信息丢失）

| 通用 `Job` 字段 | migration (`GlobalMigrationStateDTO`) | reindex (`BackfillResult`) |
|---|---|---|
| `Status` | `Status` | （净增：原来无状态机） |
| `Error` | `ErrorMessage` | （净增） |
| `StartedAt`/`FinishedAt` | `StartedAt`/`FinishedAt` | （净增） |
| `Phase` | （净增：见 §9.3，DTO 现无 phase） | （净增：可上报「340/1200」） |
| 主键 `Type` | 常量 `"migration"` | 常量 `"reindex"` |
| `Payload` (JSON) | `{version, source_type, source_payload}` | `{total, indexed, skipped, failed}` |

两客户领域字段都不大，一个 `text` JSON 列足够，**不为每个 domain 拆表**。`Payload` 用 `string`(JSON) 而非 `datatypes.JSON`——我们只整存整取、不在 SQL 里查 JSON 内部，string 最轻（开放问题 D 已定）。

## 6. 接口

### 6.1 Service / Runner / Repository

```go
// internal/job/job.go
package job

type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusSuccess   Status = "success"
    StatusFailed    Status = "failed"
    StatusCancelled Status = "cancelled"
)

// Runner（边界：擦除/untyped）——每种作业只关心「怎么干活」。
// payload 为原始 JSON；result 作为终态 Job.Payload 落库；返回 error 则置 failed。
// 必须尊重 ctx 取消（Cancel 时 ctx.Done() 触发）。
type Runner interface {
    Run(ctx context.Context, payload []byte, report ReportFunc) (result any, err error)
}

// ReportFunc：Runner 上报进度（仅进内存，§8）。phase 必填，snapshot 可为 nil。
type ReportFunc func(phase string, snapshot any)

type JobRepository interface {
    // Upsert 按主键 Type 原子写入/覆盖。
    Upsert(ctx context.Context, j *Job) error
    GetByType(ctx context.Context, jobType string) (Job, error)   // 查无返回 (zero, ErrNotFound)
    // SweepRunning：启动期把残留的 pending/running 行批量置 failed（§8 / 开放问题 B）。
    SweepRunning(ctx context.Context, reason string) error
}

type Service struct {
    runners map[string]Runner
    repo    JobRepository
    live    map[string]*Progress       // 按 type；仅当前在跑作业的实时进度，内存态
    cancels map[string]context.CancelFunc
    mu      sync.Mutex
}

func (s *Service) Register(jobType string, r Runner)                       // 启动期注册
func (s *Service) Submit(jobType string, payload []byte) (Job, error)      // 互斥→upsert pending→go run
func (s *Service) Get(jobType string) (Job, error)                         // durable ⊕ 内存进度
func (s *Service) Cancel(jobType string) error
```

> 接口是 **type-centric**（不是 id-centric）：因每 type 单行，访问路径就是 type；前端轮询直接打 `/reindex/status`，无需传 id。

### 6.2 泛型边界 `Adapt[P]`（作者端 typed，注册表 untyped）

异构注册表 `map[type]Runner` 装着不同 payload 类型的 Runner，Go 泛型无法让它们共存于一个 map——**边界必然擦除**。因此泛型只放作者端：

```go
type TypedRun[P any] func(ctx context.Context, p P, report ReportFunc) (any, error)

func Adapt[P any](fn TypedRun[P]) Runner {
    return runnerFunc(func(ctx context.Context, raw []byte, report ReportFunc) (any, error) {
        var p P
        if len(raw) > 0 {
            if err := json.Unmarshal(raw, &p); err != nil {
                return nil, fmt.Errorf("decode %T payload: %w", p, err)
            }
        }
        return fn(ctx, p, report)
    })
}

// 注册（reindexRunner.Run 直接拿 ReindexPayload，零 map[string]any 脏活）：
svc.Register("reindex",   job.Adapt(reindexRunner.Run))
svc.Register("migration", job.Adapt(migrationRunner.Run))
```

reindex 无输入 payload（`ReindexPayload struct{}`）；migration 的 `MigrationPayload{ SourceType string; SourcePayload map[string]any }` 照样能放下现有 map，不丢信息。

## 7. 状态机与生命周期

### 7.1 正常流转

```
Submit ─► pending ─(go run 起协程)─► running ─┬─► success   (Runner 返回 nil err)
                                              ├─► failed    (Runner 返回 err)
                                              └─► cancelled (ctx 被 Cancel)
```

- `Submit`：校验 type 已注册 → §7.3 互斥校验 → upsert `Job{Type, Status:pending}` → **同步登记 cancelFunc 进 `cancels[type]`（§7.2）** → 起 goroutine → 立即返回该 `Job`。
- goroutine 内：置 `running` + `StartedAt`（落库）→ 调 `runner.Run(ctx, payload, report)` → 据返回值落终态 `Status` + `FinishedAt` + 终态 `Phase`/`Payload`（落库）→ 清理 `live[type]`/`cancels[type]`。
- `report(phase, snapshot)`：**只更新内存 `live[type]`，不落库**（§8）。

### 7.2 取消（消除 pending 竞态）

- `Submit` 内**同步**（返回前、持锁）把 `context.WithCancel` 的 `cancel` 存入 `cancels[type]`。由此**不存在「pending 已建、cancelFunc 未登记」的窗口**——`Cancel` 永远能命中非终态作业，pending/running 走同一条路（开放问题 C 已定）。
- `Cancel(type)`：取 `cancels[type]` 调用 → Runner 的 `ctx.Done()` 触发 → Runner 协作退出 → goroutine 落 `cancelled`。若已终态/无在跑作业，则 no-op。
- Runner **必须**在长循环里检查 `ctx.Err()`（reindex 的 page 循环、migration 的 ETL 阶段间）。

### 7.3 同类型互斥

每 `type` 同时只允许一条非终态作业。`Submit` 先 `GetByType`，若现存行 `Status ∈ {pending, running}` 则报错（沿用 migration 现有「请先结束/清理当前迁移」语义）。`Type` 主键 + 单事务内「读判断 + upsert」保证原子，无需全局锁。

## 8. 进度持久化策略：durable 落库，progress 只进内存

fine-grained 进度（如 reindex「340/1200」）**无持久价值**——进程一旦重启，该 run 必被 §8/B 扫成 failed，进度作废。因此持久化它纯属浪费写（reindex 大库逐页 `report` 会变成数百次 `UPDATE` 砸 SQLite 单写者）。

策略（对标 Sidekiq/Resque「进度进 Redis、durable 进 DB」，单进程场景用内存替 Redis）：

- **落 DB（durable，每 job 全程约 3~4 次，与库大小无关）**：submit→pending、start→running、terminal→success/failed/cancelled + 终态 Phase/Payload。
- **进内存（ephemeral）**：`report` 写 `live[type]`，永不碰 DB。

```go
func (s *Service) report(jobType, phase string, snapshot any) {     // 仅内存
    s.mu.Lock(); s.live[jobType] = &Progress{Phase: phase, Snapshot: snapshot}; s.mu.Unlock()
}

func (s *Service) Get(jobType string) (Job, error) {               // durable 打底 ⊕ 叠加实时进度
    row, err := s.repo.GetByType(ctx, jobType)
    if err != nil { return row, err }
    if p := s.live[jobType]; p != nil {                            // 本进程正在跑
        row.Phase, row.Payload = p.Phase, mustJSON(p.Snapshot)
    }
    return row, nil
}
```

优点：无定时器、无「丢最后一次更新」边界；与 §8 孤儿清理、§5.2 单行 upsert 天然咬合。**仅当**未来真要「进度跨重启续显」（明确不要）才退回节流落库。

**启动期孤儿清理（开放问题 B 已定）**：`job.Service` 初始化时 `repo.SweepRunning("interrupted by restart")`，把上次进程残留的 `pending/running` 行置 `failed`。`live` 随进程蒸发 → `Get` 返回 failed，语义自洽，避免前端永久转圈。幂等、零成本。

## 9. 迁移映射（migration & reindex）

### 9.1 ReindexRunner

把 `EmbeddingService.Backfill` 的 page 循环包成 Runner：
- `Run` 复用现有循环；每页结束 `report("indexing", BackfillResult{...})` 上报累计计数（仅内存）。
- 循环里加 `if ctx.Err() != nil { return ... }` 支持取消。
- 终态 result = `BackfillResult`（落 `Payload`）。
- 倾向「Runner 调 service」：`Backfill` 保留为 service 方法，仅加 `report`/`ctx` 取消支持，循环主体不下沉。

### 9.2 ⚠️ 契约坑一：`idle` 哨兵

migration 现在「没在跑」时返回 `status: "idle"`（`MigrationStatusIdle`），前端轮询依赖它判断「无进行中迁移」。通用 `Job` **没有 idle**——「没在跑」= `GetByType("migration")` 返回 `ErrNotFound`。

处理：迁移时，migration 的 status 查询端点在 `GetByType` 查无行时**合成一个 `idle` 响应**，保持前端契约不破。这层「Job ↔ 旧 DTO」适配放在 migrator service/handler，不污染通用 `job.Service`。

### 9.3 ⚠️ 契约坑二：migration 现在没填 Phase

`job.go` 定义了 `MigrationPhaseExtracting/...` 常量，但 `GlobalMigrationStateDTO` **不带 phase 字段**——迁移现在前端看不到细粒度阶段。通用 `Job.Phase` 是**净增能力**：迁移时让 `MigrationRunner` 在 ETL 各阶段 `report(phase, …)`，进度条免费升级。reindex 同理。

### 9.4 删除的手写状态机

`internal/service/migrator/migrator.go` 里随 §9 一并删除/改写：
- `activeMu` / `activeCancel`（取消改走 `jobService.Cancel`）
- `saveGlobalStateWithRetry` / `runGlobalMigration` / `getGlobalState`（状态读写改走 `job.Service`/`repo`）
- `MigrationGlobalJobStateKey`（KeyValue 键）退役（老库残留键无害，忽略）。

## 10. DI 接入点

- 新增 `internal/job` provider：`NewService`（注入 `JobRepository`）、`NewJobRepository`、各 `Runner` 构造。
- `Runner` 依赖领域 service（`ReindexRunner`←`EmbeddingService`，`MigrationRunner`←迁移 ETL），按现有分层别名规则注入（`jobService` 等）。
- Runner 注册：一个 `RegisterRunners(svc, reindexRunner, migrationRunner)` provider，或在 `app` 组件 Start 时注册。
- **set 归属**：`job.Service` 同时被 `HandlerSet`（HTTP 触发）和 `TaskerSet`（定时触发）需要，应放进共享的 `InfraSet`/`DomainSet`，避免 wire 为两个 Build 各生成一个实例——参考 `internal/di/wire.go:40` 的 `VisitorSet` 注释（同款坑：两个 Build 各生成一个 Tracker 导致读写不同实例）。
- 改动后 `just wire` 重新生成 `wire_gen.go`，CI 跑 `just wire-check`。

## 11. 前端轮询契约

- reindex：新增 `GET /api/.../reindex/status`（按 type，无需 id）返回 `status + phase + payload`，前端渲染进度条与计数。reindex 触发端点改为「起 job 即返回」，不再阻塞。
- migration：响应形状**保持现状**（含 `idle`，§9.2），前端不改；用 `phase` 渲染细粒度阶段是增量增强。
- 复用 migration 既有轮询组件/范式，不引入新轮询机制。

## 12. 落地分期

- **PR1：框架 + reindex**。建 `internal/job`（`Service`/`Runner`/`Adapt`/`JobRepository`/`Job` 表）、`ReindexRunner`、reindex handler 改异步 + status 端点、前端 reindex 轮询、DI 接线、`just wire`。migration **暂不动**（PR1 不碰它，短期并存可接受）。
- **PR2：迁 migration**。`MigrationRunner`、删手写状态机、`idle` 哨兵适配、删 `MigrationJob` 死表、phase 上报增强、migration 回归测试。

> §2.1-G3 要求最终「无两套状态机并存」，PR1↔PR2 间的短暂并存是过渡，不是终态。

## 13. 风险与回归测试

- **migration 回归**（PR2 最大风险）：start/status/cancel 全链路、`idle` 哨兵、互斥（重复 start 报错）、success/failed/cancelled 四态、tmp 清理行为不退化。
- **reindex 超时消除**：大库（构造 N 千条 Echo）下不再阻塞 HTTP；取消能中断 page 循环。
- **崩溃语义**：进程在 `running` 中途重启 → 启动期 `SweepRunning` 把残留行扫成 `failed`，前端不再永久转圈（§8）。
- **进度写入压力**：验证大库 reindex 期间 DB 写次数为常数级（≈ 生命周期转换数），不随页数增长（§8）。
- **死表 drop**：不主动 drop `migration_jobs`，留空表无害（§5.1）。

## 14. 决策记录（原开放问题 A–E，已定）

- **A. `ReportFunc` 签名** → `report(phase string, snapshot any)`，**带 snapshot**。reindex 需上报实时计数；配合 §8 进度纯内存，无写压力顾虑。
- **B. 孤儿 running 清理** → **做**。启动期 `SweepRunning` 置 `failed`（§8）。
- **C. Cancel 掉 pending** → **消除 pending 窗口**：`Submit` 内同步登记 cancelFunc，pending/running 同路可取消（§7.2）。
- **D. Payload 类型** → `string`(JSON)，不用 `datatypes.JSON`（§5.3）。
- **E. backup-export** → 维持非目标，但设计须保证它仅是「再加一个 Runner + 一个 type 常量」即可接入；当前接口满足。

---

_主要决策已收敛。下一步进入 PR1（框架 + reindex）。_
