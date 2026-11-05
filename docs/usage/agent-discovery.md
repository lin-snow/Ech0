# 智能体发现（llms.txt / AI catalog / MCP server card）

> 面向部署者与集成方：让 AI 智能体在**不握手协议**的前提下发现本实例的订阅源、API 与 MCP 端点。
> 关联文档：[`usage/mcp-usage.md`](mcp-usage.md)（MCP 接入与授权）、[`dev/huma-raw-gin-endpoints.md`](../dev/huma-raw-gin-endpoints.md)（为什么这四个端点走裸 gin）。

## 1. 四个文件

| 路径 | 类型 | 标准 | 作用 |
|---|---|---|---|
| `/llms.txt` | `text/plain`（Markdown） | [llmstxt.org](https://llmstxt.org/) | 人类可读的入口清单：站点、订阅源、API 文档、MCP 卡片 |
| `/.well-known/ard.json` | `application/ai-catalog+json` | [ARD 1.0](https://agenticresourcediscovery.org/spec/) | 条目清单；`entries[].url` 指向下面的 server card |
| `/.well-known/ai-catalog.json` | `application/ai-catalog+json` | ARD 前身路径（等价条目源） | 同一份文档；Lighthouse 与既有 AI Catalog 消费方仍只认这条路径 |
| `/mcp/server-card` | `application/mcp-server-card+json` | [MCP SEP-2127](https://modelcontextprotocol.io/seps/2127-mcp-server-cards) | MCP 连接前置信息：传输、协议版本、鉴权头 |

三者关系：

```text
robots.txt  --Agentmap-->  /.well-known/ard.json  --entries[].url-->  /mcp/server-card
                                  |                                          |
                                  +----------- 同一个 MCP 端点 /mcp ----------+
```

`/llms.txt` 只是把这些绝对 URL 连同订阅源、OpenAPI 文档一起列出来。

## 2. 内容边界

四份文档**只描述实例，不描述内容**：

- 不含帖文、评论或用户数据，因此私有实例、邀请制实例同样可以安全开启；
- 每条链接都是绝对 URL，来源是**请求的对外 origin**（`X-Forwarded-Proto` / `X-Forwarded-Host` 优先），与 `/robots.txt`、`/sitemap.xml` 一致；
- MCP 卡片只声明「怎么连」，不声明工具清单：工具随 token 的 scope 变化，属于运行时信息（`GET /api/mcp/manifest`）。

## 3. 开关

偏好设置 → 系统设置 → **智能体发现**，默认开启。

关闭后：

- `/llms.txt`、`/.well-known/ard.json`、`/.well-known/ai-catalog.json`、`/mcp/server-card` 一律返回 `404`；
- `/robots.txt` 去掉 `Agentmap:` 行（`Sitemap:` 保留）。

存储字段是 `disable_agent_discovery`（**语义取反**）：写入该字段之前的设置 JSON 里没有这个键，反序列化得到 `false`，即「开启」——升级不会悄悄关掉这个能力。需要整站拒绝 AI 抓取时，仍应在反向代理或 CDN 层拦截，而不是只靠这个开关：它只撤销发现入口，不改 `robots.txt` 的抓取许可。

## 4. 校验

上线后逐条确认：

```bash
curl -sS https://your.instance/llms.txt | head -5          # 应为 "# <站点标题>"
curl -sSI https://your.instance/.well-known/ard.json       # 200 + application/ai-catalog+json + ETag
curl -sS https://your.instance/robots.txt | grep Agentmap  # 应为 "Agentmap: https://your.instance/.well-known/ard.json"
curl -sS https://your.instance/mcp/server-card | jq .remotes  # type=streamable-http, url=<origin>/mcp
```

反向代理若没有转发 `X-Forwarded-*`，文档里的绝对 URL 会退化成内网地址——这与 `/sitemap.xml` 的既有要求一致，配置方式见 [`usage/configuration.md`](configuration.md)。

浏览器侧可用 DevTools → Lighthouse → *Agentic browsing* 分类复核：`llms.txt follows recommendations` 与 `ai-catalog.json schema is valid` 两项应通过。
