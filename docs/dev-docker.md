# 本地开发：Docker 依赖（Postgres + Redis）

本机 **不要安装** PostgreSQL / Redis，统一使用仓库根目录 `docker-compose.yml` 启动容器。

## 启动

```bash
# 仓库根目录
docker compose up -d
docker compose ps
```

健康检查通过后：

| 服务 | 主机地址 | 默认账号 |
|------|----------|----------|
| PostgreSQL | `localhost:5432` | 用户/密码 `postgres` / `postgres`，库名 `cntuanyuan` |
| Redis | `localhost:6379` | 无密码 |

首次空数据卷会自动执行 `backend/migrations/postgres/00_bootstrap.sql` 建表与种子数据。  
**已有 volume 不会重复执行**；改 schema 请用迁移脚本或初始化向导。

## 后端连接

复制示例配置或使用初始化向导，关键参数示例：

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  database: "cntuanyuan"
  ssl_mode: "disable"

redis:
  host: "localhost"   # 空字符串表示不使用 Redis
  port: 6379
```

```bash
cd backend
go run cmd/app/main.go
# 或
go run cmd/app/main.go --config ./config/config.yaml
```

## Web

```bash
cd web
npm install
npm run dev
```

## 常用命令

```bash
docker compose logs -f postgres
docker compose logs -f redis
docker compose restart
docker compose down          # 停容器，保留数据卷
docker compose down -v       # 停容器并删除数据卷（会清空库）
```

## 说明

- 仅依赖进容器；Go 后端与 Next.js 管理端在本机运行。
- 生产部署仍参考根目录 `DEPLOY.md`（二进制 + Nginx 等），本 compose 面向本地/联调。
