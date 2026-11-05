# 环境变量与 `.env` 配置参考

面向部署者：Ech0 读哪些环境变量、默认值是什么，以及 env 与管理后台设置的边界。
为什么要分两层、为什么有些项必须留在 env，见 [配置与设置的依赖架构](../dev/config-and-settings-architecture.md)。

## 加载顺序与 `.env`

启动时按以下优先级取值（高 → 低，按 key 逐个合并）：

1. **进程环境**：`docker -e`、Compose `environment`、systemd `Environment=`、shell `export`。
2. **工作目录下的 `./.env`**：由 `godotenv.Load()` 读取，**不会**向上级目录查找。
3. **`internal/config/config.go` 内置默认值**。

行为细节：

- 没有 `--env-file` / `ECH0_ENV_FILE`；要换文件只能换工作目录。Docker 镜像的 `WORKDIR` 是 `/app`，因此挂载 `/app/.env` 即生效，但更常见的是 Compose 的 `env_file:`。
- 没有 `.env` 是正常情况，不再打印提示。文件存在但无法解析时，stderr 打印 `Failed to load .env: ...`，随后继续使用进程环境 + 默认值，**不会阻止启动**。
- 数值变量写错（例如 `ECH0_UPLOAD_IMAGE_MAX_SIZE=20MB`）时，stderr 打印 `Invalid environment overrides (...)`；出错的字段保留默认值，其余字段照常生效。
- 未识别的变量（包括拼错的 `ECH0_xxx`）被静默忽略，不会报错。
- 配置只在启动时解析一次（`sync.Once`）；改 env 或 `.env` 后必须重启进程。

```bash
# 最小 .env 示例（放在工作目录）
JWT_SECRET=please-change-me
ECH0_SERVER_PORT=6277
ECH0_DB_PATH=data/ech0.db
```

## 两层配置：env 改不动已有的后台设置

| | Bootstrap 配置 | 运行时设置 |
|---|---|---|
| 来源 | env / `.env` / 内置默认值 | KV 设置表（管理后台写入） |
| 生效时机 | 进程启动 | 保存即生效 |
| 典型项 | 端口、DB、日志、JWT 密钥、事件/Agent 运行时参数 | 站点信息、评论、S3、OAuth2、Passkey、Agent 供应商 |

`ECH0_SETTING_*`、`ECH0_S3_*`、`ECH0_AUTH_WEBAUTHN_*` 这类变量只在
**对应设置第一次落库前**（全新数据库，或手动删除了该 KV key）充当默认值 —— `setting.Seed` 只写缺失的 key，绝不覆盖管理后台
已保存的值。已经初始化过的实例，改这些 env 重启也不会生效，请到后台对应页面修改。

少数项是「env 直读」而非种子，改了在下次启动时生效：端口/监听地址、DB、日志、JWT、上传大小上限、
`ECH0_EVENT_*`、`ECH0_AGENT_*`、验证码参数、`ECH0_WEB_CORS_ALLOWED_ORIGINS`（MCP 的 Origin 放行）。
`ECH0_AUTH_REDIRECT_ALLOWED_RETURN_URLS` 介于两者之间：既是 OAuth2 设置首次落库的默认值，也在设置页白名单为空时
作为运行时回退。

## 变量一览

默认值与 `internal/config/config.go` 中 `defaultConfig()` 一致；`(空)` 表示空字符串或空列表。

### 服务与文档

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_SERVER_PORT` | `6277` | HTTP 监听端口 |
| `ECH0_SERVER_HOST` | `0.0.0.0` | 监听地址；空、`0.0.0.0`、`::` 均表示所有网卡（保持双栈）。设为 `127.0.0.1` 等可限制来源 |
| `ECH0_SERVER_MODE` | `release` | `debug` 时开启 Gin 请求日志、控制台彩色日志 |
| `ECH0_OPENAPI_DOCS_RENDERER` | `stoplight` | `/api/docs` 渲染器：`stoplight`（CDN）或 `scalar`（离线内置） |
| `ECH0_EGRESS_ALLOW_PRIVATE` | `false` | 关闭出站 SSRF 守卫：放行 `localhost`、私网/CGNAT（含 Tailscale `100.64/10`）、保留地址。**仅本地开发**；`just run` / `just dev` 已注入 `true` |

> `ECH0_EGRESS_ALLOW_PRIVATE=true` 对所有出站调用生效（connect、webhook、OAuth CIMD、站点抓取），
> 等于把 SSRF 面重新打开。只在不可被外部访问的开发实例上使用；生产环境保持 `false`，
> 需要连通内网 peer 时优先走公网/隧道域名。


### 数据库

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_DB_TYPE` | `sqlite` | **仅支持 `sqlite`**；其他值启动即失败（避免空 DB 句柄） |
| `ECH0_DB_PATH` | `data/ech0.db` | SQLite 文件路径，文件名任意；父目录会自动创建 |
| `ECH0_DB_LOGMODE` | `release` | `release` 关闭 GORM 日志；其他值仅输出 error 级 |

### 日志

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` / `panic` / `fatal` |
| `ECH0_LOG_FORMAT` | `json` | `json` 或 `console` |
| `ECH0_LOG_CONSOLE` | `true` | 是否写 stdout；设为 `false` 后只写文件与 TUI 日志流（`docker logs` 将看不到日志） |
| `ECH0_LOG_FILE_ENABLE` | `true` | 是否写日志文件 |
| `ECH0_LOG_FILE_PATH` | `data/app.log` | 日志文件路径（Lumberjack 轮转） |
| `ECH0_LOG_FILE_MAX_SIZE` | `100` | 单文件上限（MB） |
| `ECH0_LOG_FILE_MAX_BACKUPS` | `5` | 保留的历史文件数 |
| `ECH0_LOG_FILE_MAX_AGE` | `30` | 历史文件保留天数 |
| `ECH0_LOG_FILE_COMPRESS` | `true` | 轮转后 gzip |
| `ECH0_LOG_BUFFER_SIZE` | `2048` | 日志流缓冲区条数 |
| `ECH0_LOG_RECENT_SIZE` | `2000` | TUI/日志流保留的最近条数 |
| `ECH0_LOG_DROP_POLICY` | `drop_oldest` | 缓冲满时 `drop_oldest` 或 `drop_newest` |
| `ECH0_LOG_FLUSH_BATCH` | `128` | 文件写入批量条数 |
| `ECH0_LOG_FLUSH_INTERVAL_MS` | `500` | 文件写入刷新间隔（ms） |

### 鉴权

| 变量 | 默认值 | 说明 |
|---|---|---|
| `JWT_SECRET` | `(空)` | **注意没有 `ECH0_` 前缀**。也可用 `ECH0_JWT_SECRET`（别名，前者优先）。不设置则每次启动随机生成，**重启后所有已签发 token 失效**，stderr 会给出警告 |
| `ECH0_JWT_EXPIRES` | `900` | access token 有效期（秒） |
| `ECH0_JWT_REFRESH_EXPIRES` | `604800` | refresh token 有效期（秒） |
| `ECH0_JWT_ISSUER` | `ech0` | JWT `iss` |
| `ECH0_JWT_AUDIENCE` | `ech0` | JWT `aud` |
| `ECH0_AUTH_REDIRECT_ALLOWED_RETURN_URLS` | `(空)` | 逗号分隔；OAuth2 白名单：设置页为空时回退到它，首次落库也取它作默认 |
| `ECH0_AUTH_WEBAUTHN_RP_ID` | `(空)` | Passkey 设置首次落库的默认 RP ID |
| `ECH0_AUTH_WEBAUTHN_ORIGINS` | `(空)` | 逗号分隔；Passkey 默认允许来源 |

### 上传

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_UPLOAD_IMAGE_MAX_SIZE` | `20971520` | 图片单文件上限（字节，20 MiB） |
| `ECH0_UPLOAD_AUDIO_MAX_SIZE` | `20971520` | 音频单文件上限（20 MiB） |
| `ECH0_UPLOAD_VIDEO_MAX_SIZE` | `67108864` | 视频单文件上限（64 MiB） |

> 允许的 MIME 类型是代码内置常量，不可通过 env 修改；存储子目录由扩展名路由决定
> （`images/`、`audios/`、`videos/`、`documents/`、`files/`），也不是 env 可配项。

### 存储

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_STORAGE_DATA_ROOT` | `data/files` | 本地存储根目录，同时决定 `/api/files` 静态路由 |
| `ECH0_OBJECT_ENABLED` | `false` | S3 设置首次落库的默认开关 |
| `ECH0_S3_ENDPOINT` | `(空)` | S3 兼容端点（不带 `http(s)://`） |
| `ECH0_S3_ACCESS_KEY` | `(空)` | 访问密钥 |
| `ECH0_S3_SECRET_KEY` | `(空)` | 私钥 |
| `ECH0_S3_BUCKET` | `(空)` | 桶名 |
| `ECH0_S3_REGION` | `(空)` | 区域 |
| `ECH0_S3_PROVIDER` | `(空)` | `aws` / `r2` / `minio` / `other` |
| `ECH0_S3_USE_SSL` | `false` | 是否使用 HTTPS |
| `ECH0_S3_USE_PATH_STYLE` | `false` | 是否使用 path-style 寻址 |
| `ECH0_S3_CDN_URL` | `(空)` | 对外媒体域名 |
| `ECH0_S3_PATH_PREFIX` | `(空)` | 对象 key 前缀（无首尾斜杠） |

> `ECH0_S3_*` 只是 S3 设置首次落库的默认值；之后一律以后台「存储设置」为准。
> 迁移规则见 [storage-migration.md](storage-migration.md)。

### 事件运行时（Busen）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_EVENT_DEFAULT_BUFFER` | `512` | 默认订阅缓冲条数 |
| `ECH0_EVENT_DEFAULT_OVERFLOW` | `block` | `fail_fast` / `drop_newest` / `drop_oldest`；其他值按 `block` 处理 |
| `ECH0_EVENT_SYSTEM_BUFFER` | `64` | 系统事件（顺序消费）缓冲 |
| `ECH0_EVENT_AGENT_BUFFER` | `128` | Agent 事件缓冲 |
| `ECH0_EVENT_AGENT_PARALLELISM` | `2` | Agent 订阅并发度 |
| `ECH0_EVENT_WEBHOOK_POOL_WORKERS` | `6` | Webhook 工作协程数 |
| `ECH0_EVENT_WEBHOOK_POOL_QUEUE` | `6` | Webhook 任务队列长度 |

### Agent（Copilot）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_AGENT_TIMEOUT_SECONDS` | `300` | 单次生成预算（秒），模型流式输出 + 非交互工具；`<=0` 表示不限时 |
| `ECH0_AGENT_MAX_ROUNDS` | `20` | 单次运行最多工具调用轮数 |
| `ECH0_AGENT_ASK_TIMEOUT_SECONDS` | `300` | 等待用户回答问题/确认写操作的最长时间（秒） |

### 评论验证码（内置 gocap）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_COMMENT_CAPTCHA_SITE_KEY` | `ech0-comment` | 站点 key，决定 `/api/cap/<key>/` 端点 |
| `ECH0_COMMENT_CAPTCHA_SECRET` | `(空)` | 未设置时由内部密钥派生 |
| `ECH0_COMMENT_CAPTCHA_DIFFICULTY` | `4` | 工作量证明难度 |
| `ECH0_COMMENT_CAPTCHA_CHALLENGE_COUNT` | `80` | 每个挑战的分片数 |
| `ECH0_COMMENT_CAPTCHA_SALT_SIZE` | `32` | 盐长度 |
| `ECH0_COMMENT_CAPTCHA_CHALLENGE_TTL` | `900` | 挑战有效期（秒） |
| `ECH0_COMMENT_CAPTCHA_REDEEM_TTL` | `7200` | 兑换凭证有效期（秒） |
| `ECH0_COMMENT_CAPTCHA_GC_INTERVAL` | `2` | 过期数据清理间隔（秒） |
| `ECH0_COMMENT_CAPTCHA_ENABLE_CORS` | `true` | 验证码端点是否允许跨源 |
| `ECH0_COMMENT_CAPTCHA_IP_HEADER` | `(空)` | 反向代理场景下取真实 IP 的头名 |
| `ECH0_COMMENT_CAPTCHA_MAX_BODY_BYTES` | `1048576` | 请求体上限（字节） |
| `ECH0_COMMENT_CAPTCHA_RATE_LIMIT_MAX` | `30` | 限流次数上限 |
| `ECH0_COMMENT_CAPTCHA_RATE_LIMIT_WINDOW` | `5` | 限流窗口（秒） |
| `ECH0_COMMENT_CAPTCHA_RATE_LIMIT_SCOPE` | `cap` | 限流维度 |
| `ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_REDEEM` | `false` | 兑换接口是否参与限流 |
| `ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_SITEVERIFY` | `false` | 校验接口是否参与限流 |

> 评论开关、是否强制验证码等属于运行时设置，在后台「评论管理」页面配置；**没有** `ECH0_COMMENT_ENABLE`
> 这类 env（早期遗留的同名变量已移除，设置了也不会生效）。

### 站点设置种子项

以下变量只在对应设置第一次落库时生效（见上文「两层配置」）。之后的修改请走管理后台。

| 变量 | 默认值 |
|---|---|
| `ECH0_SETTING_SITE_TITLE` | `Ech0` |
| `ECH0_SETTING_SERVER_LOGO` | `/Ech0.svg` |
| `ECH0_SETTING_SERVER_NAME` | `Ech0` |
| `ECH0_SETTING_SERVER_URL` | `https://ech0.example.com` |
| `ECH0_SETTING_ALLOW_REGISTER` | `true` |
| `ECH0_SETTING_ICP_NUMBER` | `(空)` |
| `ECH0_SETTING_FOOTER_CONTENT` | `(空)` |
| `ECH0_SETTING_FOOTER_LINK` | `(空)` |
| `ECH0_SETTING_METING_API` | `(空)` |
| `ECH0_SETTING_CUSTOM_CSS` | `(空)` |
| `ECH0_SETTING_CUSTOM_JS` | `(空)` |

### 其他

| 变量 | 默认值 | 说明 |
|---|---|---|
| `ECH0_WEB_CORS_ALLOWED_ORIGINS` | `(空)` | 逗号分隔；MCP `/mcp` 的额外跨源白名单（直读），同时作为 OAuth2 设置首次落库的默认值 |

## 常见坑

1. **`JWT_SECRET` 没有前缀。** 写成 `ECH0_JWT_SECRET` 现在也可用；两者都没设则每次重启换密钥，表现为「所有客户端突然需要重新登录」。
2. **改 `ECH0_SETTING_*` 不生效。** 该实例已经落库过设置，env 只是首次默认值。请到后台改，或删掉对应 KV key 后重启。
3. **`docker logs` 空。** 镜像内 `ECH0_LOG_CONSOLE` 默认 `true`；若被显式设为 `false`，日志只进 `data/app.log` 与 TUI 日志流。
4. **自定义 `ECH0_DB_PATH`。** 文件名任意（如 `data/mydb.sqlite`），父目录会自动创建；空值会让启动失败。
5. **`.env` 放错位置。** 只读当前工作目录，`/opt/ech0/.env` 对在 `/opt/ech0/data` 启动的进程无效。
