# deploy/docker

## 作用
容器化相关配置说明。

## 本地依赖（PostgreSQL + Redis + MinIO）

**镜像**：Compose 使用 `pull_policy: never`，**只使用本机已有镜像**，不会从仓库拉取。默认镜像名为 `postgres:13`、`redis:latest`、`minio/minio:latest`（与常见本机环境一致）；若你本地是其他标签，在本目录复制 `.env.example` 为 `.env`，设置 `POSTGRES_IMAGE`、`REDIS_IMAGE`、`MINIO_IMAGE`（与 `docker images` 中 `REPOSITORY:TAG` 一致）。

首次使用 MinIO 前需在本机拉取镜像（示例）：

```bash
docker pull minio/minio:latest
```

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

默认端口：`5432`（PostgreSQL）、`6379`（Redis）、`9000`（MinIO S3 API）、`9001`（MinIO 控制台）。数据持久化在命名卷 `vehivle_pg_data`、`vehivle_redis_data`、`vehivle_minio_data`。

**MinIO（本地模拟 OSS / S3）**：浏览器访问 `http://localhost:9001`，使用 compose 中的 `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`（当前默认为 `vehivle` / `vehivle123`）登录，在控制台中创建 Bucket。应用侧使用 S3 兼容 SDK 时，典型配置为：

- Endpoint：`http://localhost:9000`（容器内访问 API 服务名时为 `http://minio:9000`）
- Region：可填 `us-east-1`（MinIO 常用占位）
- Access Key / Secret Key：与上述根账号一致
- 使用路径样式（path-style / force path style）可避免部分 SDK 对虚拟主机风格的兼容问题

若生产环境为阿里云 OSS，请在联调时区分 endpoint 与签名算法差异；本地 MinIO 仅用于开发联调。

## 建议内容
- API 服务 Dockerfile。
- 依赖服务编排：`docker-compose.yml`（已提供开发用 PostgreSQL / Redis）。

## 注意
镜像应区分构建阶段和运行阶段，尽量减小体积并提高安全性。
