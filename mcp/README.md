# Echo MCP 服务使用说明

| ![r1nj7tOlAVY49CN](https://s2.loli.net/2025/12/07/r1nj7tOlAVY49CN.png) | ![gRJ6uA8YODl5dHS](https://s2.loli.net/2025/12/07/gRJ6uA8YODl5dHS.png) |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| ![VTZLzOGN6CvkafW](https://s2.loli.net/2025/12/07/VTZLzOGN6CvkafW.png) | ![48qWZQjFRcliwSY](https://s2.loli.net/2025/12/07/48qWZQjFRcliwSY.png) |



## 目录

1. [简介](#简介)
2. [环境要求](#环境要求)
3. [快速开始指南](#快速开始指南)
4. [安装与配置](#安装与配置)
5. [本地使用](#本地使用)
6. [线上部署](#线上部署)
7. [AI客户端配置](#ai客户端配置)
8. [端口约定与说明](#端口约定与说明)
9. [API操作说明](#api操作说明)
10. [认证方式详解](#认证方式详解)
11. [高级用法](#高级用法)
12. [常见问题](#常见问题)

## 简介

Echo MCP服务是一个基于Model Context Protocol (MCP)协议的中间件，它允许AI助手通过标准化的方式与Echo笔记系统进行交互。通过MCP协议，您可以在任何支持MCP的AI客户端中执行对笔记系统的各种操作，包括搜索、发布、更新和删除等。

## 环境要求

- Node.js 18.0 或更高版本
- Docker (可选，用于运行Echo服务)
- 支持MCP协议的AI客户端

## 快速开始指南

1. **部署Echo服务**（如果尚未部署）：

   ```bash
   docker run -d \
     --name ech0 \
     -p 6277:6277 \
     -v /opt/ech0/data:/app/data \
     -v /opt/ech0/backup:/app/backup \
     -e JWT_SECRET="Hello Echos" \
     sn0wl1n/ech0:latest
   ```

2. **设置MCP服务器**（在主机上）：

   ```bash
   # 进入mcp目录
   cd /path/to/ech0/mcp
   
   # 安装依赖
   npm install
   
   # 启动MCP服务器（如在AI客户端配置则可以忽略启动命令，AI客户端会自动运行启动）
   npm start
   ```

3. **配置AI客户端**（以Cherry Studio为例）：

   ```json
   {
     "mcpServers": {
       "ech0": {
         "command": "env",
         "args": [
           "NOTE_HOST=http://localhost:6277", //改为你的Echo服务地址
           "NOTE_TOKEN=你的后台TOKEN",
           "node",
           "/path/to/ech0/mcp/server.js" //改为你的本地文件实际路径
         ]
       }
     }
   }
   ```

4. **开始使用**：

   - "搜索今天发布的内容"
   - "发布一条新笔记：今天学习了Echo系统"
   - "获取前5条Echo内容"
   - "删除包含'旧信息'的笔记"

### 常用场景

- **内容管理**：快速添加、编辑和删除笔记
- **知识检索**：通过自然语言搜索历史记录
- **工作流集成**：将笔记功能集成到AI助手的日常工作流中
- **自动化处理**：结合其他MCP工具实现内容处理的自动化

## 安装与配置

### 1. 获取MCP文件

```bash
# 进入mcp目录
cd /path/to/ech0/mcp

# 查看目录内容
ls
```

您应该看到以下文件：
- `server.js` - MCP服务器主文件
- `package.json` - Node.js项目配置
- `README.md` - 项目说明文档

### 2. 安装依赖

```bash
npm install
```

这将安装必要的依赖：
- `@modelcontextprotocol/sdk` - MCP官方SDK
- `zod` - 类型验证库

## 本地使用

### 1. 配置环境变量

创建一个`.env`文件或直接在命令行中设置环境变量：

```bash
# Echo服务地址（Docker运行时默认为6277端口）
export NOTE_HOST=http://localhost:6277

# 可选：设置认证令牌
export NOTE_TOKEN=你的后台token

# 可选：启用HTTP调试端口
export NOTE_HTTP_PORT=1315
```

### 2. 启动MCP服务器

```bash
npm start
```

### 3. 本地测试

您可以使用curl测试MCP服务器是否正常工作：

```bash
# 获取可用工具列表
curl http://localhost:1315/mcp/tools

# 测试工具调用
curl -N -X POST http://localhost:1315/mcp/tool/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"test","page":1,"pageSize":10}'
```

## 线上部署

### 1. Docker部署（MCP服务器）

```bash
# 构建Docker镜像
docker build -t ech0-mcp .

# 运行容器（使用令牌认证，连接到本地Echo服务）
docker run --rm -e NOTE_HOST=http://host.docker.internal:6277 -e NOTE_TOKEN=你的token ech0-mcp

# 运行容器（使用会话认证，连接到本地Echo服务）
docker run --rm -e NOTE_HOST=http://host.docker.internal:6277 ech0-mcp

# 运行容器（连接到远程Echo服务）
docker run --rm -e NOTE_HOST=https://your-echo-domain.com -e NOTE_TOKEN=你的远程token ech0-mcp
```

### 2. 同时运行Echo服务和MCP服务器

```bash
# 启动Echo服务
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest

# 构建MCP Docker镜像（需要先从源码构建）
cd /path/to/ech0/mcp
docker build -t ech0-mcp .

# 启动MCP服务器
docker run -d \
  --name ech0-mcp \
  --link ech0:ech0 \
  -e NOTE_HOST=http://ech0:6277 \
  -e NOTE_TOKEN=你的token \
  -p 1315:1315 \
  ech0-mcp
```

注意：
1. 在Docker网络中，容器之间可以通过容器名（如`ech0`）直接通信，因此使用`http://ech0:6277`作为`NOTE_HOST`

### 3. 在主机上运行MCP服务器（推荐）

```bash
# 启动Echo服务
docker run -d \
  --name ech0 \
  -p 6277:6277 \
  -v /opt/ech0/data:/app/data \
  -v /opt/ech0/backup:/app/backup \
  -e JWT_SECRET="Hello Echos" \
  sn0wl1n/ech0:latest

# 在主机上启动MCP服务器
cd /path/to/echo/mcp
npm install
export NOTE_HOST=http://localhost:6277
export NOTE_TOKEN=你的token
npm start
```

### 4. 实际使用示例（基于您的配置）

Cherry studio AI客户端配置：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "env",
      "args": [
        "NOTE_HOST=http://localhost:6277", //改为你的Echo服务地址
        "NOTE_TOKEN=你的后台TOKEN",
        "node",
        "你的本地路径/mcp/server.js"
      ]
    }
  }
}
```

### 2. Docker Compose部署

在项目根目录的`docker-compose.yml`中添加MCP服务：

```yaml
services:
  echo:
    # ... 现有配置

  echo-mcp:
    build: ./mcp
    environment:
      - NOTE_HOST=http://echo:8080
      - NOTE_TOKEN=${NOTE_TOKEN}
    ports:
      - "1315:1315"  # 可选：仅用于调试
    depends_on:
      - echo
```

启动服务：

```bash
docker-compose up -d
```

## AI客户端配置

以下是在不同AI客户端中配置Echo MCP服务器的JSON格式：

### 1. Claude Desktop

在`~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）或相应配置文件中添加：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "node",
      "args": [
        "/path/to/echo/mcp/server.js"
      ],
      "env": {
        "NOTE_HOST": "http://localhost:6277",
        "NOTE_TOKEN": "你的后台token"
      }
    }
  }
}
```

### 2. Cherry Studio

在MCP配置中添加：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "env",
      "args": [
        "NOTE_HOST=http://localhost:6277",
        "NOTE_TOKEN=你的后台token",
        "node",
        "/path/to/echo/mcp/server.js"
      ]
    }
  }
}
```

### 3. Cursor

在`.cursor/mcp_servers.json`中添加：

```json
{
  "servers": {
    "ech0": {
      "command": "node",
      "args": ["/path/to/echo/mcp/server.js"],
      "cwd": "/path/to/echo/mcp",
      "env": {
        "NOTE_HOST": "http://localhost:6277",
        "NOTE_TOKEN": "你的后台token"
      }
    }
  }
}
```

### 4. VS Code（使用MCP扩展）

在VS Code设置中添加：

```json
{
  "mcp.servers": {
    "ech0": {
      "command": "node",
      "args": ["/path/to/echo/mcp/server.js"],
      "env": {
        "NOTE_HOST": "http://localhost:6277",
        "NOTE_TOKEN": "你的后台token"
      }
    }
  }
}
```

### 5. 远程连接配置

对于远程Echo服务，修改`NOTE_HOST`：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "node",
      "args": ["/path/to/echo/mcp/server.js"],
      "env": {
        "NOTE_HOST": "https://your-echo-domain.com",
        "NOTE_TOKEN": "你的远程服务token"
      }
    }
  }
}
```

### 6. Docker环境配置

如果您的Echo服务和MCP服务器都运行在Docker中，需要调整配置：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "docker",
      "args": [
        "exec",
        "-i",
        "-e", "NOTE_HOST=http://ech0:6277",
        "-e", "NOTE_TOKEN=你的token",
        "ech0-mcp",
        "node",
        "/app/mcp/server.js"
      ]
    }
  }
}
```

## API操作说明

Echo MCP服务器提供以下工具，您可以通过AI助手调用这些工具：

### 1. 搜索Echo内容

- **工具名称**: `search` 或 `搜索`
- **无需认证**: 是
- **参数**:
  - `keyword` (可选): 搜索关键词
  - `page` (可选): 页码，默认1
  - `pageSize` (可选): 每页大小，默认10
  - `format` (可选): 输出格式，设为"json"返回原始数据

**使用示例**:
```
请搜索包含"技术分享"的Echo内容
```

### 2. 获取分页内容

- **工具名称**: `page` 或 `页面`
- **无需认证**: 是
- **参数**:
  - `page` (可选): 页码，默认1
  - `pageSize` (可选): 每页大小，默认10

**使用示例**:
```
获取第2页的Echo内容，每页显示5条
```

### 3. 获取指定Echo

- **工具名称**: `echo`
- **无需认证**: 是
- **参数**:
  - `id` (必需): Echo的ID

**使用示例**:
```
获取ID为123的Echo内容
```

### 4. 发布新Echo

- **工具名称**: `publish` 或 `发布` 或 `笔记` 或 `说说`
- **需要认证**: 是
- **参数**:
  - `content` (必需): Echo内容
  - `type` (可选): 类型，默认"text"，可选"markdown", "image", "multipart"
  - `private` (可选): 是否私有，默认false
  - `image` 或 `images` (可选): 图片URL或URL数组
  - `imageURL` (可选): 单个图片URL（兼容性）

**使用示例**:
```
发布一条新笔记："今天学习了Node.js的异步编程模式"
```
```
发布一条带图片的Echo，图片URL：https://example.com/image.jpg，内容：美丽的风景
```

### 5. 更新Echo

- **工具名称**: `update` 或 `更新`
- **需要认证**: 是
- **参数**:
  - `id` (必需): Echo的ID
  - `content` (必需): 更新后的内容

**使用示例**:
```
更新ID为456的Echo内容为："这是更新后的内容"
```

### 6. 删除Echo

- **工具名称**: `delete` 或 `删除`
- **需要认证**: 是
- **参数**:
  - `id` (可选): Echo的ID
  - `keyword` (可选): 用于搜索匹配的内容关键词
  - `content` (可选): 用于搜索匹配的内容片段

**使用示例**:
```
删除ID为789的Echo
```
```
删除包含"过时的信息"的Echo
```

### 7. 获取今日Echo

- **工具名称**: `today` 或 `今日`
- **需要认证**: 是
- **参数**: 无

**使用示例**:
```
获取今天发布的所有Echo内容
```

### 8. 获取标签列表

- **工具名称**: `tags` 或 `标签`
- **无需认证**: 是
- **参数**: 无

**使用示例**:
```
获取所有可用的标签
```

### 9. 获取系统状态

- **工具名称**: `status` 或 `状态`
- **无需认证**: 是
- **参数**: 无

**使用示例**:
```
检查Echo系统状态
```

### 10. 用户登录

- **工具名称**: `login` 或 `登录`
- **需要认证**: 否（用于获取会话）
- **参数**:
  - `username` (必需): 用户名
  - `password` (必需): 密码

**使用示例**:
```
使用用户名admin和密码password123登录Echo系统
```

## 认证方式

Echo MCP服务器支持两种认证方式：

### 1. 令牌认证

在环境变量中设置`NOTE_TOKEN`：

```bash
export NOTE_TOKEN=your_api_token_here
```

或者直接在AI客户端配置中设置：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "node",
      "args": ["/path/to/echo/mcp/server.js"],
      "env": {
        "NOTE_HOST": "http://localhost:8080",
        "NOTE_TOKEN": "your_api_token_here"
      }
    }
  }
}
```

### 2. 会话认证

如果不设置`NOTE_TOKEN`，您需要先调用`login`工具进行登录：

```
请使用用户名"admin"和密码"your_password"登录Echo系统
```

成功登录后，MCP服务器将自动保存会话信息用于后续请求。

## 常见问题

### 1. 连接问题

**问题**: 无法连接到Echo服务

**解决方案**:
- 检查`NOTE_HOST`环境变量是否正确设置（确保使用6277端口）
- 确认Echo服务是否正在运行：`docker ps | grep ech0`
- 检查端口映射是否正确：`docker port ech0`
- 检查网络连接和防火墙设置

### 2. 端口冲突问题

**问题**: `EADDRINUSE :::1315` 端口占用错误

**解决方案**:
- 停止其他使用1315端口的进程
- 对于多个MCP实例，将除第一个外的所有实例`NOTE_HTTP_PORT`设为`0`
- 使用不同端口号：`export NOTE_HTTP_PORT=1316`

### 3. Docker网络问题

**问题**: Docker容器内的MCP无法连接到主机上的Echo服务

**解决方案**:
- 使用`host.docker.internal`作为主机地址：`NOTE_HOST=http://host.docker.internal:6277`
- 或使用Docker网络中的容器名：`NOTE_HOST=http://ech0:6277`
- 确保容器在同一个Docker网络中

### 4. 认证问题

**问题**: 提示"需要认证"或"令牌无效"

**解决方案**:
- 确认`NOTE_TOKEN`是否有效
- 尝试重新登录获取新的会话：调用`login`工具
- 检查令牌是否已过期，在Echo系统后台重新生成令牌

### 5. SSE连接问题

**问题**: `Invalid content type, expected text/event-stream`

**解决方案**:
- 检查反向代理配置，确保正确转发`/mcp/sse`
- 确保关闭`proxy_buffering`
- 直接访问而不是通过代理：`curl -N http://localhost:1315/mcp/sse`

### 6. 工具调用问题

**问题**: 工具调用失败或返回错误

**解决方案**:
- 检查工具参数是否正确
- 确认Echo服务API是否有变更
- 查看MCP服务器日志获取详细错误信息
- 检查认证状态，某些操作需要登录

### 7. 工具名称问题

**问题**: AI客户端提示工具名称校验警告

**解决方案**:
- 优先使用英文名称：`search`, `publish`, `delete`, `update`, `echo`, `page`, `status`, `today`, `tags`, `login`
- 中文名称可用但可能触发警告：`搜索`, `发布`, `删除`, `更新`, `消息`, `页面`, `状态`, `今日`, `标签`, `登录`

### 8. 输出格式问题

**问题**: AI客户端显示原始JSON或"乱码"

**解决方案**:
- 在AI客户端提示中要求解析JSON并用自然语言描述结果
- 在MCP调用时使用`format=json`参数获取原始数据
- 调整AI客户端设置以优化输出格式

### 9. 性能问题

**问题**: 响应缓慢

**解决方案**:
- 检查网络连接质量
- 减少`pageSize`参数值
- 确认Echo服务性能是否正常：检查CPU和内存使用情况
- 考虑使用本地Echo服务而非远程服务

### 10. 调试方法

启用调试模式：

```bash
export NOTE_HTTP_PORT=1315
npm start
```

调试命令：
```bash
# 获取可用工具列表
curl http://localhost:1315/mcp/tools

# 测试特定工具
curl -N -X POST http://localhost:1315/mcp/tool/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"test","page":1,"pageSize":10}'

# 查看SSE事件流
curl -N http://localhost:1315/mcp/sse | head -n 10
```

### 11. 环境变量问题

**问题**: 环境变量未生效

**解决方案**:
- 确保环境变量名称正确（区分大小写）
- 在Windows系统上使用`set`而非`export`：`set NOTE_HOST=http://localhost:6277`
- 在AI客户端配置中直接设置环境变量（如前面示例）
- 重启AI客户端确保环境变量加载

## 端口约定与说明

以下是Echo系统和MCP服务器的端口约定，请根据实际情况调整配置：

### 端口说明
- **Echo后端API与Web界面**：`6277`（Docker运行时的默认端口，通过`-p 6277:6277`映射）
- **MCP HTTP/SSE（可选）**：`1315`（设置`NOTE_HTTP_PORT=1315`并映射`-p 1315:1315`）
- **仅Stdio握手**：`NOTE_HTTP_PORT=0`（不监听HTTP，不占用任何端口）

### 端口使用建议
- MCP作为客户端访问`NOTE_HOST`（如`http://localhost:6277`），不会与后端端口产生冲突
- 只有在开启MCP的HTTP/SSE服务时才需要`1315`端口，主要用于调试和监控
- 避免端口冲突：如需同时运行多个MCP实例，请将第二实例的`NOTE_HTTP_PORT`设为`0`

### Docker网络中的端口访问
- Docker容器间可以通过容器名直接通信，例如：`http://ech0:6277`
- 从Docker容器访问主机服务，可使用：`http://host.docker.internal:6277`

## 认证方式详解

Echo MCP服务器支持两种认证方式：

### 1. 令牌认证（推荐）

在环境变量中设置`NOTE_TOKEN`：

```bash
export NOTE_TOKEN=your_api_token_here
```

或者在AI客户端配置中设置：

```json
{
  "mcpServers": {
    "ech0": {
      "command": "node",
      "args": ["/path/to/echo/mcp/server.js"],
      "env": {
        "NOTE_HOST": "http://localhost:6277",
        "NOTE_TOKEN": "your_api_token_here"
      }
    }
  }
}
```

获取令牌的方法：
1. 访问Echo系统Web界面
2. 登录后进入设置或管理页面
3. 查找"API令牌"或"Token"相关选项
4. 生成新的令牌并复制

### 2. 会话认证

如果不设置`NOTE_TOKEN`，您需要先调用`login`工具进行登录：

```
请使用用户名"admin"和密码"your_password"登录Echo系统
```

成功登录后，MCP服务器将自动保存会话信息用于后续请求。

### 认证工具权限说明
- **无需认证**：`search`、`page`、`echo`、`status`、`tags`
- **需要认证**：`publish`、`delete`、`update`、`today`、`login`

## 高级用法

### 1. SSE事件监控

MCP服务器提供Server-Sent Events (SSE)端点，用于监控工具执行事件：

```bash
# 连接到SSE流
curl -N http://localhost:1315/mcp/sse

# 预期输出：
# event: mcp_hello
# data: {"name":"ech0-mcp","version":"0.1.0"}
# event: mcp_tools
# data: ["search","publish",...]
# event: keepalive
# data: {}
```

### 2. Docker执行方式

使用`docker exec`方式运行MCP，适合已运行的Echo容器：

```json
{
  "mcpServers": {
    "ech0-docker-exec": {
      "command": "docker",
      "args": [
        "exec",
        "-i",
        "-e", "NOTE_HOST=http://host.docker.internal:6277",
        "-e", "NOTE_HTTP_PORT=0",
        "-e", "NOTE_TOKEN=你的token",
        "ech0",
        "node",
        "/path/to/mcp/server.js"
      ]
    }
  }
}
```

### 3. 使用代理

如果您的网络环境需要代理，可以设置：

```bash
export https_proxy=http://your-proxy:port
export http_proxy=http://your-proxy:port
```

### 4. 自定义超时

您可以通过环境变量设置请求超时（毫秒）：

```bash
export REQUEST_TIMEOUT=10000
```

### 5. 日志级别

设置日志级别：

```bash
export LOG_LEVEL=debug  # 可选: error, warn, info, debug
```

## 结语

通过Echo MCP服务，您可以将Echo笔记系统无缝集成到AI助手中，实现更高效的内容管理和检索。无论您是开发者、内容创作者还是知识管理者，这个工具都能显著提高您的工作效率。如果您遇到任何问题或需要帮助，请查看本文档的"常见问题"部分

> 参考：[来源](https://github.com/rcy1314/echo-noise/blob/main/mcp/README.md)
> 该协议还可继续优化增强，前提是需要配合API来同步增强
