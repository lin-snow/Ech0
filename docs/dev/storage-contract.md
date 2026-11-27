# 存储契约（Storage Contract）

> 已废弃，仅保留作历史入口：现行契约见 [`docs/dev/file-system-design.md`](file-system-design.md)。

本文件描述的是早期"四类对象 + `Ref` 寻址"的设计。该设计从未按原样落地：`Ref` 没有成为公共寻址值，`instance / avatar / content / temp` 四类 key 命名空间与 `PublicPathPrefixes()` 公开面、`File.URL` 快照、`EchoFile` / `InstanceAsset` 关联表都已删除。取而代之的是 `files` 单表 + `kind`（owned/blob/ref）+ `file_links` 引用、`storage.Store` SPI 与唯一内容网关 `GET|HEAD /api/files/{id}/content`。继续维护两份契约只会产生分叉，因此本文被 `file-system-design.md` 整体取代；存储的布局、授权、生命周期、迁移与验收标准一律以那份文档为准。
