---
title: 统一登录
description: 第三方 OAuth、OIDC 与 Passkey 的配置与排错（新手向）
---

「统一登录」让你不必单独给 Ech0 记一套密码：可以用 **GitHub / Google / QQ** 等账号登录，或用 **Passkey（指纹、面容、安全密钥）** 无密码登录。在管理后台进入 **系统设置 → SSO** 分别配置 **OAuth2** 与 **Passkey**；两者可同时开启，用户在登录页按需选择。

---

## 这篇文档适合谁读

- 你是**站长/管理员**，要在自己的实例上打开第三方登录或 Passkey。
- 你**不需要**先懂 OAuth 协议细节；下面会先讲「发生了什么」，再给操作步骤。
- 普通用户只需在登录页点对应按钮；**绑定第三方账号**在个人设置里完成。

---

## 先搞懂两件事

### 1. OAuth2 / OIDC 在干什么（白话）

可以把流程想成「去银行柜台办联名卡」：

1. 用户在 Ech0 点「用 GitHub 登录」。
2. 浏览器**跳到 GitHub**，由用户在那里登录并**授权**「允许 Ech0 读取你的公开资料」之类权限。
3. GitHub 把用户**带回**你的 Ech0 站点，并附带一次性的**授权码**。
4. Ech0 服务器用授权码向 GitHub 换 **Token**，再拉取用户资料，在本地**创建或匹配账号**，最后让用户登录成功。

**回调地址（Redirect URI）** 就是第 3 步里「GitHub 必须把用户送回哪里」——必须和 Ech0 里显示的地址**完全一致**（协议、域名、路径、末尾斜杠都要一致），否则就会报 **Redirect URI mismatch**。

**OIDC** 可以理解为「OAuth 的升级版」：除了 access token，还会多一个 **id_token**（JWT），适合对接企业 IdP（如 Keycloak、Authentik）。在 Ech0 里打开 OIDC 相关选项后，会校验 Issuer、JWKS 等字段。

### 2. Passkey 在干什么（白话）

Passkey 用**设备上的生物识别或硬件安全密钥**代替密码：浏览器和 Ech0 之间走 **WebAuthn** 标准。  
配置里有两个容易混淆的概念：

| 配置项                        | 简单理解                                                        | 示例                       |
| ----------------------------- | --------------------------------------------------------------- | -------------------------- |
| **RP ID（WebAuthn RP ID）**   | 「哪个网站」在管这把钥匙，一般是**域名**，**不要**写 `https://` | `memo.example.com`         |
| **Origins（允许的站点来源）** | 浏览器从**哪个完整 URL** 访问时允许用这把钥匙，必须带协议       | `https://memo.example.com` |

两者必须与浏览器地址栏里**实际访问的域名一致**。换了域名、只用 IP、HTTP/HTTPS 混用，都会导致注册或登录失败。

---

## 开始之前请确认

- 你已用**管理员账号**登录，能进入 **系统设置 → SSO**。
- 生产环境建议使用 **HTTPS** 与**固定域名**（OAuth 与 Passkey 在纯 IP、自签证书环境下常出问题）。
- 在第三方平台创建应用时，需要能接收邮件或使用开发者账号（如 GitHub、Google Cloud、QQ 互联）。

---

## OAuth2：从创建应用到能登录

### 回调地址长什么样

Ech0 使用路径（注意：**没有** `/api` 前缀）：

```text
https://你的站点域名/oauth/<provider>/callback
```

其中 `<provider>` 只能是 **`github`**、**`google`**、**`qq`** 或与模板对应的 **`custom`**（自定义/OIDC 时常用）。  
管理后台 **OAuth2** 页里会显示**自动生成的 Redirect URI**，请**原样复制**到第三方平台，**不要手改路径或多加斜杠**。

### 在 Ech0 里要填什么

1. 在第三方平台创建 OAuth 应用，拿到 **Client ID** 和 **Client Secret**。
2. 打开 **系统设置 → SSO → OAuth2**，启用开关，选择模板：**GitHub / Google / QQ / 自定义**。
3. 把页面上的 **Redirect URI** 填回第三方平台。
4. 把 **Client ID / Secret** 填进 Ech0；**Scope（权限）**按「够用即可」填写，例如：
   - GitHub：`read:user`（读取公开资料）
   - Google：常填 `openid`、`email`、`profile`（以当前控制台要求为准）
5. 若使用 **OIDC**（自定义 IdP），在界面中补全 **Issuer、JWKS URL** 等；保存后服务端会对 **id_token** 做校验。

### 各平台创建应用（入口与注意点）

**GitHub**

1. 打开 GitHub → **Settings → Developer settings → OAuth Apps → New OAuth App**。
2. **Authorization callback URL** 填：`https://你的域名/oauth/github/callback`（与后台显示一致）。
3. 创建后记录 **Client ID**，生成 **Client Secret** 并妥善保存。

**Google**

1. 使用 [Google Cloud Console](https://console.cloud.google.com/) 创建项目，启用 **Google+ API / 用户信息服务**（以控制台当前名称为准）。
2. 在「凭据」里创建 **OAuth 2.0 客户端 ID**，应用类型选 **Web**。
3. **已授权的重定向 URI** 填：`https://你的域名/oauth/google/callback`。

**QQ**

1. 在 [QQ 互联](https://connect.qq.com/) 创建网站应用，通过审核后获取 **APP ID / APP Key**。
2. **回调地址**填：`https://你的域名/oauth/qq/callback`（与 Ech0 后台一致）。

**自定义 / OIDC**

1. 选择 **自定义** 模板，按 IdP 文档填写 **授权 URL、Token URL、用户信息 URL**。
2. 开启 OIDC 时填写 **Issuer、JWKS**，权限范围按 IdP 要求填写。

### 「授权回跳」与 CORS（什么时候要改）

若你从**多个前端域名**访问同一套后端（例如 `https://a.com` 与 `https://b.com`），可能需要在 OAuth 高级设置里配置 **Auth Redirect Allowed Return URLs** 与 **CORS Allowed Origins**，否则登录成功后的**跳转**或**跨域请求**会被拒绝。  
一般**单域名自建**使用默认值即可；修改后保存，并确保与反向代理、环境变量中的站点地址一致。

### 绑定到已有账号

用户**已用密码等方式登录**后，可在 **个人设置** 里将 GitHub/Google 等身份**绑定**到当前账号，避免重复建号。

---

## Passkey：配置步骤与常见坑

### 配置顺序（建议）

1. **系统设置 → SSO → Passkey**。
2. 填写 **WebAuthn RP ID**（你的域名，无协议）与 **WebAuthn Origins**（如 `https://你的域名`，可多条）。
3. 保存后界面应显示 **Passkey 就绪** 一类状态。
4. 在支持 WebAuthn 的浏览器中**注册** Passkey（系统会提示指纹、面容或插入安全密钥）。
5. 同一页可**管理**已注册设备（重命名、删除）。

### 常见错误

| 现象                    | 可能原因                                                                      |
| ----------------------- | ----------------------------------------------------------------------------- |
| 无法注册 / 登录 Passkey | RP ID 与当前访问域名不一致，或 Origins 未包含当前完整 URL                     |
| 仅 IP 访问              | WebAuthn 通常要求**安全上下文**（HTTPS 或 localhost）；公网建议用域名 + HTTPS |
| 换了域名后全部失效      | 需用新域名重新配置 RP ID / Origins，并让用户重新注册 Passkey                  |

需要对接 HTTP API 时，可打开部署实例上的 **`/swagger/index.html`** 查看 Passkey 相关接口说明。

---

## OAuth 与 Passkey 一起用

可以同时启用：用户既可用第三方账号登录，也可在登录后绑定 Passkey，下次用通行密钥登录。具体按钮与文案以当前版本登录页为准。

---

## 排错速查

| 报错或现象                | 处理方向                                                                          |
| ------------------------- | --------------------------------------------------------------------------------- |
| **Redirect URI mismatch** | 第三方填的回调与 Ech0 显示的**逐字符一致**；检查 `http`/`https`、末尾 `/`、子域名 |
| **Invalid client**        | Client ID/Secret 复制错误，或应用未启用、未完成审核                               |
| OAuth 能跳回但无法登录    | 检查 Scope 是否包含拉取用户信息所需权限；OIDC 时检查 Issuer/JWKS                  |
| Passkey 无响应            | 浏览器是否支持；站点是否 HTTPS；RP ID/Origins 是否与地址栏一致                    |

更细的接口与字段说明以你部署实例上的 **OpenAPI（Swagger）** 为准。
