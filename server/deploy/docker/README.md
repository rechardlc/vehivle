# deploy/docker

## 作用
容器化相关配置说明。

## 本地依赖（PostgreSQL + Redis）

**镜像**：Compose 使用 `pull_policy: never`，**只使用本机已有镜像**，不会从仓库拉取。默认镜像名为 `postgres:13`、`redis:latest`（与常见本机环境一致）；若你本地是其他标签，在本目录复制 `.env.example` 为 `.env`，设置 `POSTGRES_IMAGE`、`REDIS_IMAGE`（与 `docker images` 中 `REPOSITORY:TAG` 一致）。

在本目录启动：

```bash
docker compose up -d
```

查看状态：

```bash
docker compose ps
```

停止（保留数据卷）：

```bash
docker compose down
```

在 `server/` 下配置数据库与 Redis（与 `configs/conf.dev.yaml`、`.env.example` 一致），示例：

```env
VEHIVLE_DATABASE_DSN=postgres://vehivle:vehivle@localhost:5432/vehivle?sslmode=disable
VEHIVLE_REDIS_ADDR=localhost:6379
```

默认端口：`5432`（PostgreSQL）、`6379`（Redis）。数据持久化在命名卷 `vehivle_pg_data`、`vehivle_redis_data`。

## 建议内容
- API 服务 Dockerfile。
- 依赖服务编排：`docker-compose.yml`（已提供开发用 PostgreSQL / Redis）。

## 注意
镜像应区分构建阶段和运行阶段，尽量减小体积并提高安全性。
