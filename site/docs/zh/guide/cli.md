---
title: 命令行（CLI）
description: 在终端发帖、审核评论、查看实例
---

`ech0` 是 Ech0 的命令行客户端，适合批量操作、脚本和服务器排障。日常使用在网页里点也完全可以。

---

## 安装

```bash
npm install -g @useech0/cli
```

支持 macOS / Linux（arm64 / x64），自带运行时，不需要额外装 Node 或 Go；暂不支持 Windows。也可以从 [GitHub Releases](https://github.com/lin-snow/Ech0/releases) 下载 `ech0-cli-<os>-<arch>.tar.gz`。

---

## 登录

```bash
ech0 auth login --server https://你的域名 --device
```

浏览器打开提示的链接并确认即可，令牌默认保存 1 个月。登录信息按实例分开保存，多台实例用 `ech0 config list` / `ech0 config use <名字>` 切换。

其他登录方式（脚本或 CI）：

```bash
ech0 auth login --server https://你的域名 -u admin          # 用户名 + 隐藏输入密码
echo "$TOKEN" | ech0 auth login --server https://你的域名 --token -
```

CI 里也可以直接用 `ECH0_SERVER` + `ECH0_TOKEN` 两个环境变量，不落盘。

---

## 先看看这台实例

```bash
ech0 info
```

一次显示：站点名、版本、帖子数、是否已初始化、当前登录状态、命令清单是否最新。`--json` 输出给脚本。

---

## 常用操作

```bash
# 浏览
ech0 echo list --page 1 --page-size 20
ech0 echo get --id <帖子ID>

# 发帖（图片先上传，再引用返回的 ID）
ech0 file upload --file photo.png
ech0 echo create --content "今天的想法" --tag daily --file-id <文件ID>

# 评论审核
ech0 comment panel list --status pending
ech0 comment panel status --id <评论ID> --status approved

# 互联健康
ech0 connect health
```

- 删除类命令会先确认；脚本或 `--json` 输出必须加 `--yes`。
- 文件没有删除命令：文件可能仍被帖子引用，请在网页面板处理。

---

## 权限

默认登录已覆盖上面这些命令。如果只想给脚本只读权限，登录时指定：

```bash
ech0 auth login --server https://你的域名 --device --scopes echo:read,file:read
```

调用返回 403 就是令牌缺权限，重新登录换一个令牌即可。完整权限说明见 [访问令牌](/zh/docs/guide/accesstoken)。

---

## 帮助与输出

```bash
ech0 help                       # 所有命令
ech0 help echo                  # 某一组
ech0 help comment panel list    # 单条命令的字段与示例
```

`--json` 输出原始响应，`--yaml` 输出 YAML。退出码：`0` 成功，`2` 用法错误，`1` 服务端或网络错误。

服务端新增接口不需要升级 CLI：命令清单会随实例更新，过期时自动刷新。

---

## 给 AI 助手用

```bash
ech0 install skills ~/.claude/skills
```

把 Ech0 的命令说明（含权限与安全约定）装进 Claude Code 等工具的 skills 目录。
