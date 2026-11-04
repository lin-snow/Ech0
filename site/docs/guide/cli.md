---
title: 命令行（CLI）
description: 安装、登录、实例信息、常用命令与权限
---

`ech0` 是 Ech0 的命令行客户端：登录到一台实例，在终端完成 Echo、文件、标签、评论（含审核）、互联、Webhook 与实例信息查询。命令树由实例下发的 **operation catalog** 生成，服务端新增接口不需要升级 CLI。

---

## 安装

```bash
npm install -g @useech0/cli
# 或者
bun add -g @useech0/cli
```

- 按平台分发（macOS / Linux × arm64 / x64），自带运行时，不需要 Node、Bun 或 Go；**暂不支持 Windows**。
- 也可以从 [GitHub Releases](https://github.com/lin-snow/Ech0/releases) 下载 `ech0-cli-<os>-<arch>.tar.gz`（校验文件为同页的 `SHA256SUMS`）。

---

## 登录

```bash
ech0 auth login --server https://你的域名 --device          # 浏览器设备码（推荐）
ech0 auth login --server https://你的域名 -u admin          # 用户名 + 隐藏输入密码
echo "$TOKEN" | ech0 auth login --server https://你的域名 --token -
```

- 管理员登录会铸造一枚 `audience=cli`、有效期 **1 个月** 的访问令牌；普通用户保存会话令牌，过期时自动续期。
- 凭据按 **profile** 保存（系统钥匙串优先，回退到 `0600` 文件）；`ECH0_NO_KEYCHAIN=1` 可强制用文件。多实例用 `ech0 config list` / `ech0 config use <name>` 切换。
- `--token -` 从 stdin 读令牌，避免写进 shell history；CI 里也可直接用 `ECH0_SERVER` + `ECH0_TOKEN`。

---

## 查看实例

```bash
ech0 info
```

一次只读汇总：站点名、服务地址、版本、Echo 总数与今日数、初始化 / owner 状态、当前 token 来源与有效性、catalog 指纹与 op 数。它聚合的是公开端点，**未登录也能用**；如果本地 catalog 缓存落后于服务端，会提示 `ech0 catalog refresh`。加 `--json` 给脚本用。

```bash
ech0 auth status --verify   # 只看凭据：profile、token 来源，并用 /api/user 校验一次
```

---

## 常用命令

```bash
# 时间线
ech0 echo list --page 1 --page-size 20 --json
ech0 echo list --search daily --sort-by fav_count --json
ech0 echo get --id <echo_id> --json
ech0 echo hot / today / random / onthisday --json

# 发帖（附件先上传，再引用 file id）
ech0 file upload --file ./photo.png --json
ech0 echo create --content "配图" --tag daily --file-id <file_id>

# 评论：发布 + 审核闭环
ech0 comment create --echo-id <echo_id> --content "写得好"
ech0 comment panel list --status pending --json      # 需要 comment:review
ech0 comment panel status --id <comment_id> --status approved
ech0 comment panel delete --id <comment_id> --yes    # 需要 comment:delete

# 互联
ech0 connect list --json
ech0 connect info --json
ech0 connect health --json                           # 需要 connect:read
ech0 connect add --connect-url https://peer.example

# 标签与 Webhook
ech0 tag list --json
ech0 webhook list --json
```

- 删除类命令（`echo delete`、`tag delete`、`comment panel delete`、`connect delete`、`webhook delete`）默认交互确认；管道、CI 与 `--json`/`--yaml` 必须显式加 `--yes`。
- `ech0 file delete` **不提供**：文件可能仍被 Echo 引用，终端删除的影响面不可见，请在面板处理。

---

## 权限（Scope）

登录时会按命令面申请默认 scope：`echo:read/private/write/delete`、`profile:read`、`file:read/write`、`comment:write/review/delete`、`connect:read/write/delete`、`admin:webhook`。需要更窄的令牌时显式传入：

```bash
ech0 auth login --server https://你的域名 --device --scopes echo:read,file:read
```

命令所需 scope 直接来自 catalog，并在 `--help` 里显示；`403` 表示令牌缺该 scope，需要重新登录或在面板重新铸造。完整词表见 [访问令牌](/docs/guide/accesstoken)。

---

## 命令树与帮助

```bash
ech0 help                      # 本地命令 + 服务端命令地图
ech0 help echo                 # 某一组的命令
ech0 help comment panel list   # 单条命令的字段、示例与 scope
ech0 help --json               # 给脚本 / agent 的机器版本
ech0 catalog refresh           # 重新拉取命令清单（服务端提示过期时会自动重试）
```

命令清单缓存在本地；服务端接口变化后，CLI 收到 `412` 会自动刷新并重试一次。

---

## 输出与退出码

- `--json` 原样打印响应体；`--yaml` 输出同一文档；流式接口在 `--json` 下每行一个事件。
- 退出码：`0` 成功、`2` 用法错误（缺参数 / 非法取值）、`1` 服务端或网络错误。
- 错误写 stderr，并附带一行可执行的提示。

---

## 给 AI agent 用

```bash
ech0 install skills ~/.claude/skills
```

把内置的 Ech0 skill（`SKILL.md` + `references/*.md`）写进 skills 根目录，让 Claude Code 等工具知道有哪些命令、scope 与安全约束；`--from <folder>` 可安装本地版本。
