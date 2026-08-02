# Ech0 胶囊（Capsule）使用说明

这份文档基于当前项目实现编写，目标是让你可以直接把内容导出、迁移，并编译成一个能白嫖静态托管的只读站点。

规范性定义见 [`../dev/capsule/spec.md`](../dev/capsule/spec.md)，设计取舍见 [`../dev/capsule/capsule-design.md`](../dev/capsule/capsule-design.md)。

---

## 1. 胶囊是什么

**胶囊是一个自包含的目录（或 zip），装着一个 Ech0 站点的全部公开内容，人类可读、可编辑、可用 git 管理。**

```text
capsule/
  ech0.yaml         # 站点信息 + 归属 + 互联列表
  echoes/2026/2026-08-02-b302099a.md   # 一条 Echo = 一个 frontmatter-markdown 文件
  comments.yaml     # 评论快照（公开投影，不含 email / IP）
  files/images/…    # 媒体字节，和实例本地存储同构
```

它解决三件事：

- **搬家**：把内容从一个实例整体搬到另一个实例，不碰数据库文件。
- **长期保存**：十年后没有 Ech0，这些 markdown 和图片照样能读。
- **静态发布**：编译成纯静态站，扔到 GitHub Pages / Cloudflare Pages 就是一个只读的个人站。

### 胶囊 vs 快照

| | 胶囊 capsule | 快照 snapshot |
|---|---|---|
| 内容 | 公开内容（Echo / 媒体 / 评论 / 站点设置） | 整个实例（含数据库文件、账号、凭据、运维配置） |
| 形态 | 可读目录，逐条 markdown | 不透明 zip |
| 导入语义 | **幂等合并**，按 id 跳过，不覆盖 | **破坏性替换**整库 |
| 用途 | 搬家 / 备份内容 / 建静态站 / 换工具 | 灾备回滚 |

要备份「整台机器」用快照；要带走「我写的东西」用胶囊。

---

## 2. 命令一览

所有命令都在**实例根目录**下执行（即 `data/` 所在目录）。

```bash
ech0 export capsule   [-o ./capsule] [--include-private] [--zip]
ech0 export snapshot  [-o ./snapshot.zip]
ech0 import capsule   [<path>=./capsule] [--include-private] [--dry-run]
ech0 import snapshot  <snapshot.zip> --yes
ech0 check            [<path>=./capsule] [--fix]
ech0 build            [<path>=./capsule] [-o ./dist] [--base-url /]
```

退出码：`0` 成功，`1` 校验错误或执行失败。警告不影响退出码。

---

## 3. 导出

```bash
ech0 export capsule -o ./my-capsule
```

- 默认**排除**私密 Echo。要带走加 `--include-private`（胶囊不加密，别随手往公开仓库推）。
- 媒体是**记录驱动**导出：以 `files` 表为准，S3 上的对象会被下载进胶囊。
- **自包含是硬承诺**：只要有一个托管文件的字节取不回来（S3 不可达、对象丢了），导出就整体失败并列出清单，绝不产出一个「看起来完整、实际缺图」的胶囊。修好再跑即可，失败不会在输出目录留残档。
- `--zip` 产出单文件；zip 内布局与目录形态完全一致。

外链文件（`storage_type=external`）只带 URL，字节不随胶囊走——那些字节本来就不归你管。

---

## 4. 校验

```bash
ech0 check ./my-capsule          # 目录或 .zip 都行
ech0 check ./my-capsule --fix
```

分两级：**错误**会让 import / build 拒绝执行；**警告**只提示。

常见警告：

- `embeds source instance URL` — 正文或 logo 里写死了原实例地址，迁移后可能断链。v1 不自动改写（保「逐字一致」契约），你自己决定要不要改。
- `dangling media` — 胶囊里有字节，但既没有 Echo 引用它、清单 `files` 块也没声明它。`ech0 export` 不会产生这种情况（未挂 Echo 的附件会进清单），基本只在手写胶囊里出现。
- `custom_js / custom_css is not empty` — 别人给的胶囊里带着脚本，导入等于执行对方代码。看一眼再说。

`--fix` **只**补一件事：缺失的 `id`（生成 UUIDv7 回写 frontmatter）。手写胶囊时很有用。zip 形态不可写，`--fix` 会直接拒绝。

---

## 5. 导入

```bash
ech0 import capsule ./my-capsule --dry-run   # 先看清单，不写库
ech0 import capsule ./my-capsule
```

落地规则，记住这几条就够：

- **幂等**：按 `id` 判重，已存在的 Echo 直接跳过，不覆盖不合并。重复执行安全，没有 `--overwrite`。
- **原样入库**：`username`、`fav_count`、正文一律逐字，不做任何转换。唯一的例外是补全内部必填的用户外键——按 `username` 找同名用户挂接，找不到就挂到 owner 名下，但**展示的作者名始终是胶囊里写的那个**。
- **站点设置只填空位**：你已经配过的项一个都不会被动；从没配过的（还是默认值）才用胶囊的值填上——所以「搬到新实例」能把站点标题、logo、备案号一并带过去。
- **评论**跟着宿主 Echo 走：库里有那条 Echo 就挂上（这轮新建的、上次导入的都算），没有就记为孤儿跳过。
- **不发事件**：导入不会触发 webhook、不会重放历史内容给订阅者。代价是向量索引不会自动跟进——导完去后台点一次索引重建。

`--dry-run` 的计数与真实运行完全一致（它照常走一遍事务再回滚），可以放心用来预演。

---

## 6. 编译静态站

```bash
ech0 build ./my-capsule -o ./dist
ech0 build ./my-capsule -o ./dist --base-url /blog/    # 部署到子路径
```

产物是一个可以直接扔进任何静态托管的目录：

```text
dist/
  index.html  404.html  assets/…   # 内嵌的 Ech0 前端，原样复用
  dataset.json                     # 烘焙好的全部数据
  api/files/…                      # 媒体
  api/connect                      # Connect 名片，别的 Ech0 实例可以来连你
  rss.xml  sitemap.xml
```

**不需要装 Node 或 pnpm** —— 前端产物已经内嵌在 `ech0` 二进制里。

静态站是**冻结展示**的：点赞数、评论都按导出时的状态只读呈现，发布 / 回复入口自动隐藏。互动痕迹是内容史的一部分，藏掉只会让存档站显得比原站「死」。

部署到 GitHub Pages / Cloudflare Pages 这类平台时，记得把 **SPA 深链兜底**指到 `404.html`，否则直接访问 `/echo/<id>` 会 404。

---

## 7. 手写与第三方胶囊

胶囊的格式是公开规范，你可以手写，也可以写个转换器把别家的数据变成胶囊。最小可用的一条 Echo：

```markdown
---
id: 0198f0a0-0000-7000-8000-000000000001
created_at: 2026-08-02T10:00:00Z
---
正文写在这里，逐字入库。
```

`id` 缺了可以用 `ech0 check --fix` 生成。时间接受任意合法 RFC3339 偏移（`+08:00` 也行），导出时会统一成 UTC。

引用媒体时，`key` 是扁平文件名，字节放在 `files/` 下按扩展名分好的目录里（`.png` → `files/images/`，其余见 spec §6）：

```yaml
files:
  - key: cover.png
    category: image
```

写完先 `ech0 check` 一遍再导入。

---

## 8. 常见问题

**导出报错说某个文件取不回来？**
S3 配置变了或对象被删了。修好存储配置再跑；实在拿不回来的，把对应的 `files` 表记录清掉。

**导入后站点标题没变？**
说明目标实例已经配过标题了——导入不覆盖你配过的东西。要换就去后台改。

**静态站图片全裂？**
检查是不是漏了 `api/files/` 目录，或者部署到子路径时没带 `--base-url`。

**能用 git 管理胶囊吗？**
可以，这是设计目标之一。Echo 是独立文件，评论单独一个文件，所以加评论不会污染内容文件的 diff。
