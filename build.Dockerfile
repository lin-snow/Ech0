# =================== 前端构建阶段 ===================
FROM node:25-alpine AS frontend-builder

WORKDIR /web

COPY web/package.json web/pnpm-lock.yaml ./

# 启用 corepack 并使用项目声明的 pnpm 版本
RUN corepack enable

# 安装依赖
RUN pnpm install --frozen-lockfile

# 复制前端源代码
COPY web/ .

# 构建前端
RUN pnpm run build --mode production

# =================== 后端构建阶段 ===================
FROM golang:1.26.2-alpine AS backend-builder

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .
COPY --from=frontend-builder /template/dist /app/template/dist

ARG TARGETOS
ARG TARGETARCH

# 构建后端二进制文件 - 使用静态链接
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -tags netgo \
    -ldflags="-linkmode external -extldflags '-static' -w -s" \
    -o ech0 ./main.go

# =================== 最终镜像 ===================
FROM alpine:latest

WORKDIR /app
ENV TZ=Asia/Shanghai

# 创建必要的目录
RUN mkdir -p /app/data /app/backup /app/template

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/ech0 /app/ech0

# 设置权限
RUN chmod +x /app/ech0

# 暴露端口
EXPOSE 6277
EXPOSE 6278

# 启动命令
ENTRYPOINT ["/app/ech0"]
CMD ["serve"]