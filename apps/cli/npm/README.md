# @ech0/cli

Ech0 的命令行客户端：登录到一台 Ech0 实例，在终端完成 Echo 的增删改查。命令树由实例下发的
operation catalog 生成，服务端新增接口不需要升级 CLI。

## 安装

```bash
npm install -g @ech0/cli
# 或者
bun add -g @ech0/cli
```

安装包按平台分发（`darwin` / `linux` × `arm64` / `x64`），自带运行时，不需要 Bun 或 Go。
不支持 Windows。

## 使用

```bash
ech0 auth login --server https://your-instance.example.com --device
ech0 echo create --content "hello"
ech0 echo delete --id <id> --yes   # 删除前默认要求确认；脚本里显式 --yes
ech0 help
```

## 其他分发渠道

独立二进制从 [GitHub Releases](https://github.com/lin-snow/ech0/releases) 下载
`ech0-cli-<os>-<arch>.tar.gz`，校验文件为同一页面的 `SHA256SUMS`。

## 链接

- 文档：[`docs/dev/cli-design.md`](https://github.com/lin-snow/ech0/blob/main/docs/dev/cli-design.md)
- 源码：[`apps/cli`](https://github.com/lin-snow/ech0/tree/main/apps/cli)
- 许可：AGPL-3.0-or-later
