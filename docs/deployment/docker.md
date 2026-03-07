# Docker 部署指南

## 快速开始

### 1. 构建镜像

```bash
cd backend
docker build -t cntuanyuan-api:latest .
```

### 2. 使用 Docker Compose 部署

```bash
docker-compose up -d
```

## Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/app/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./main"]
```

## Docker Compose 配置

```yaml
version: '3.8'

services:
  api:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=cntuanyuan
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis
    networks:
      - cntuanyuan-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=cntuanyuan
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backend/migrations/postgres:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    networks:
      - cntuanyuan-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - cntuanyuan-network

volumes:
  postgres_data:
  redis_data:

networks:
  cntuanyuan-network:
    driver: bridge
```

## 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `DB_HOST` | 数据库主机 | localhost |
| `DB_PORT` | 数据库端口 | 5432 |
| `DB_USER` | 数据库用户 | postgres |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名 | cntuanyuan |
| `REDIS_HOST` | Redis主机 | localhost |
| `REDIS_PORT` | Redis端口 | 6379 |
| `REDIS_PASSWORD` | Redis密码 | - |
| `JWT_SECRET` | JWT密钥 | - |
| `LOG_LEVEL` | 日志级别 | info |

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f api

# 停止服务
docker-compose down

# 重建并启动
docker-compose up -d --build

# 执行数据库迁移
docker-compose exec api ./main migrate
```
