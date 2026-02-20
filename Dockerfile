# =================== 构建阶段 ===================
FROM alpine:latest AS builder

WORKDIR /app
ENV TZ=Asia/Shanghai

# 获取构建平台信息
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# 创建数据目录
RUN mkdir -p /app/data

# 创建备份目录
RUN mkdir -p /app/backup

# 创建数据目录(embed版无需手动创建template)
# RUN mkdir -p /app/data && \
#     mkdir -p /app/template

# 将所有平台的 ech0 二进制复制进镜像(embed版无需复制前端资源)
COPY /backend-artifacts/* /tmp/

# 将所有平台的 ech0 二进制和前端资源复制进镜像
# COPY /backend-artifacts/* /tmp/
# COPY /frontend-asset/frontend.tar.gz /tmp/

# 解压对应平台的 ech0 二进制
RUN mkdir -p /app/template && \
    if [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "amd64" ]; then \
       tar -xzf /tmp/ech0-linux-amd64.tar.gz -C /tmp && mv /tmp/ech0-linux-amd64 /app/ech0; \
    elif [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "arm64" ]; then \
       tar -xzf /tmp/ech0-linux-arm64.tar.gz -C /tmp && mv /tmp/ech0-linux-arm64 /app/ech0; \
    elif [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "arm" ] && [ "$TARGETVARIANT" = "v7" ]; then \
       tar -xzf /tmp/ech0-linux-armv7.tar.gz -C /tmp && mv /tmp/ech0-linux-armv7 /app/ech0; \
    else \
       echo "Unsupported platform: $TARGETOS/$TARGETARCH$TARGETVARIANT" && exit 1; \
    fi && \
    # 解压前端静态资源到 /app/template
   #  tar -xzf /tmp/frontend.tar.gz -C /app/template && \
    # 清理临时文件
    rm -rf /tmp/*

# =================== 最终镜像 ===================
# FROM debian:bookworm-slim
FROM alpine:latest

WORKDIR /app
ENV TZ=Asia/Shanghai

COPY --from=builder /app /app

RUN ls -lh /app

# 设置 ech0 二进制文件的权限
RUN chmod +x /app/ech0

EXPOSE 6277
# EXPOSE 6278 (暂不默认提供SSH服务)

ENTRYPOINT ["/app/ech0"]

CMD ["web"]