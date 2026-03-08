# 团圆寻亲系统 - 后端 Dockerfile
# 多阶段构建，最小化镜像体积

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/app/main.go

# 运行阶段
FROM scratch

# 从builder复制CA证书（用于HTTPS）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 从builder复制时区数据
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从builder复制编译好的二进制文件
COPY --from=builder /app/main .

# 创建配置目录和上传目录
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/main", "-check-db"] || exit 1

# 运行应用
ENTRYPOINT ["./main"]
