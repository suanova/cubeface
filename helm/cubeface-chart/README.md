# CubeFace Helm Chart

把 CubeFace 部署到 Kubernetes。CubeFace 是单个 Go 二进制（前端已内嵌），因此本 Chart 只有一个 Deployment 和 Service，不依赖 Redis，也不引入 CloudNativePG、redis-operator 等子 Chart。

- **SQLite 模式（默认）**：数据库文件放在 PVC 的 `/data/cubeface.db`，只能单副本运行。
- **PostgreSQL 模式**：连接外部 PostgreSQL，可水平扩容，默认带 `psql` 等待初始化容器。

## 前置条件

- Kubernetes 1.23+
- Helm 3.8+
- 可用的 CubeFace 镜像。默认 `ghcr.io/cubeface/cubeface:<appVersion>`；也可以自行构建并推送到私有仓库后覆盖 `image.repository` / `image.tag`。
- SQLite 模式需要一个默认 StorageClass，或在安装时指定 `persistence.storageClassName`。
- PostgreSQL 模式需要一个可访问的 PostgreSQL 实例；DSN 用户需要在其 schema 上拥有建表/改表权限（应用启动时执行 `AutoMigrate`）。

## 快速安装（SQLite）

```bash
helm install cubeface ./helm/cubeface-chart \
  --namespace cubeface --create-namespace
```

默认没有 Ingress，使用端口转发访问。本地端口要与 `config.PORT` 一致，否则浏览器 Origin 和 `PUBLIC_ORIGIN` 不匹配：

```bash
kubectl -n cubeface port-forward svc/cubeface 3000:80
# 打开 http://localhost:3000
```

首次安装会生成 `SESSION_SECRET` 并写入 `cubeface-secret`。请备份它：丢失后所有会话失效。

## 使用 Ingress

```bash
helm upgrade --install cubeface ./helm/cubeface-chart \
  --namespace cubeface \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set 'ingress.hosts[0]=models.example.com' \
  --set ingress.tls.enabled=true \
  --set ingress.tls.secretName=models-example-tls
```

`PUBLIC_ORIGIN` 会自动推导为 `https://models.example.com`，Session Cookie 自动加 `Secure`。Ingress 不做 rewrite，请求原样进入应用。

### 子路径部署

```bash
helm upgrade --install cubeface ./helm/cubeface-chart \
  --namespace cubeface \
  --set config.BASE_PATH=/cubeface \
  --set ingress.enabled=true \
  --set 'ingress.hosts[0]=example.com' \
  --set ingress.tls.enabled=true \
  --set ingress.tls.secretName=example-tls
```

此时站点地址是 `https://example.com/cubeface`。Ingress 只发布 `/cubeface` 前缀；API、Cookie Path、OAuth 回调和健康检查都会自动带前缀。

> 如果 TLS 在 Ingress 之外终止（例如在负载均衡器或 cert-manager 的 ssl-redirect 上），自动推导出的 `PUBLIC_ORIGIN` 可能是 `http://`。此时必须显式设置 `config.PUBLIC_ORIGIN=https://...`，否则登录会被同源校验拒绝（403 `CROSS_ORIGIN_REQUEST`）。

## 使用 PostgreSQL

```bash
helm upgrade --install cubeface ./helm/cubeface-chart \
  --namespace cubeface \
  --set database.mode=postgres \
  --set 'database.postgres.dsn=postgres://cubeface:password@db.example.com:5432/cubeface?sslmode=require' \
  --set replicaCount=2
```

- DSN 只写入 Secret，不会进入 ConfigMap。
- `database.postgres.waitForPostgres=true`（默认）会先运行 `psql -tAc "SELECT 1"`，直到数据库可连接。
- PostgreSQL 模式下 PVC 不会再被渲染，`persistence.*` 被忽略。
- 多副本时自动渲染 PDB，并给 Pod 添加软反亲和（可用 `podAntiAffinity=false` 关闭）。
- 应用启动时会执行 `AutoMigrate`。多副本同时启动会各自检查/创建表；由于迁移使用存在性检查，通常可以安全重复执行，但涉及大版本升级时建议先把 `replicaCount` 缩到 1。

## 邮箱验证与 OAuth

默认关闭邮箱验证，新装环境不需要 SMTP。开启时必须同时配置 SMTP，否则 Chart 在渲染阶段就会失败：

```bash
--set config.EMAIL_VERIFICATION_ENABLED=true \
--set config.SMTP_HOST=smtp.example.com \
--set config.SMTP_PORT=587 \
--set config.SMTP_TLS_MODE=starttls \
--set config.SMTP_USERNAME=no-reply@example.com \
--set secrets.SMTP_PASSWORD=...
```

调试模式（验证码写日志而不是发邮件）在生产环境需要显式放行：

```bash
--set config.EMAIL_DEBUG_MODE=true \
--set config.ALLOW_EMAIL_DEBUG_IN_PRODUCTION=true
```

OAuth 回调地址统一为：

```
${PUBLIC_ORIGIN}${BASE_PATH}/api/oauth/<provider>/callback
```

| 提供商 | 配置项 |
|---|---|
| GitHub | `config.GITHUB_CLIENT_ID` + `secrets.GITHUB_CLIENT_SECRET` |
| Google | `config.GOOGLE_CLIENT_ID` + `secrets.GOOGLE_CLIENT_SECRET` |
| 通用 OIDC | `config.OIDC_ISSUER_URL` / `OIDC_CLIENT_ID` / `OIDC_DISPLAY_NAME` / `OIDC_SCOPES` + `secrets.OIDC_CLIENT_SECRET` |

只有 client ID 和 secret 都配置后，提供商才会出现在登录页。未配置 secret 时对应环境变量不存在，应用会按未启用处理。

## Secret 管理

默认 `secret.create=true`：Chart 创建 `<release>-secret`，空的 `SESSION_SECRET` 会自动生成 48 位随机值；升级时通过 `lookup` 复用已有值，不会因为重新渲染而把所有人登出。其他凭据（SMTP、OAuth、初始管理员密码）不会自动生成。

使用外部 Secret 管理（External Secrets、Sealed Secrets、Vault 等）时：

```yaml
secret:
  create: false
  existingSecret: cubeface-credentials
```

该 Secret 至少需要：

| Key | 何时必需 |
|---|---|
| `SESSION_SECRET` | 总是 |
| `SQL_DSN` | `database.mode=postgres` |
| `SMTP_PASSWORD` | 使用 SMTP 时 |
| `GITHUB_CLIENT_SECRET` / `GOOGLE_CLIENT_SECRET` / `OIDC_CLIENT_SECRET` | 对应 OAuth 提供商 |
| `INITIAL_ADMIN_PASSWORD` | 使用初始化管理员时 |

其余 key 可省略。Deployment 对可选 key 使用 `optional: true`。

## 常用配置

| 值 | 默认 | 说明 |
|---|---|---|
| `image.repository` / `image.tag` | `ghcr.io/cubeface/cubeface` / Chart appVersion | 镜像地址 |
| `replicaCount` | `1` | SQLite 模式必须为 1 |
| `database.mode` | `sqlite` | `sqlite` 或 `postgres` |
| `database.postgres.dsn` | 空 | PostgreSQL DSN（写入 Secret） |
| `persistence.enabled` | `true` | SQLite 数据卷；关闭则使用 `emptyDir`（仅供测试） |
| `persistence.size` | `1Gi` | PVC 容量 |
| `persistence.storageClassName` | 空 | 不填使用集群默认 StorageClass |
| `persistence.existingClaim` | 空 | 使用已有 PVC，不再渲染新 PVC |
| `persistence.retain` | `true` | PVC 加 `helm.sh/resource-policy: keep`，卸载后保留 |
| `config.BASE_PATH` | 空 | 子路径前缀 |
| `config.PUBLIC_ORIGIN` | 自动 | 浏览器访问源，必须与真实地址一致 |
| `config.SESSION_COOKIE_SECURE` | 自动 | 空=按 `PUBLIC_ORIGIN` 是否 https 决定 |
| `config.TRUSTED_PROXIES` | 空 | 反向代理 CIDR，逗号分隔，用于取得真实客户端 IP |
| `config.REGISTER_ENABLED` | `"true"` | 是否开放注册 |
| `config.INITIAL_ADMIN_USERNAME` | 空 | 首次启动创建 root 账号的用户名 |
| `ingress.enabled` | `false` | 是否创建 Ingress |
| `service.port` | `80` | Service 端口 |
| `pdb.enabled` / `minAvailable` | `true` / `1` | 仅在副本数 > 1 时渲染 |
| `resources` | 100m/128Mi → 1/512Mi | 容器资源 |
| `envVars` | `[]` | 追加环境变量，例如 `[{name: DB_MAX_OPEN_CONNS, value: "20"}]` |
| `extraVolumes` / `extraVolumeMounts` | `[]` | 追加卷和挂载 |

完整列表和注释见 [`values.yaml`](values.yaml)。

## 健康检查

探针请求 `${BASE_PATH}/healthz`，该接口会执行数据库 `SELECT 1`。启动探针允许最多 5 分钟（`failureThreshold: 60 × 5s`），足够 SQLite 初始化或等待外部依赖。

## 校验规则

Chart 在渲染阶段会拒绝以下配置：

- `database.mode` 不是 `sqlite`/`postgres`
- SQLite 模式下 `replicaCount > 1`
- PostgreSQL 模式 + `secret.create=true` 但 DSN 为空
- `secret.create=false` 但没有 `secret.existingSecret`
- `config.BASE_PATH` 非法（含 `..`、空段或非法字符）
- `config.PORT` 非法，布尔配置不是 `true`/`false`
- 开启邮箱验证但没有 SMTP，或生产环境未放行 debug 模式
- 创建 Ingress 但没有 host，或启用 TLS 没有 `secretName`
- `config.extra` 覆盖 Chart 管理的保留键

## 升级与卸载

```bash
helm upgrade cubeface ./helm/cubeface-chart -n cubeface
helm uninstall cubeface -n cubeface
```

- SQLite 模式使用 `Recreate` 更新策略，旧 Pod 释放 RWO 卷后新 Pod 才启动，升级期间会短暂中断。
- PostgreSQL 模式使用 `RollingUpdate`（`maxUnavailable=0`），多副本可平滑滚动。
- `persistence.retain=true` 时 `helm uninstall` 不会删除 PVC 及用户数据；需要清理时手动删除 PVC：

  ```bash
  kubectl -n cubeface delete pvc cubeface-data
  ```

- 自动生成的 `SESSION_SECRET` 在卸载后随 Secret 一起删除；如果 PVC 保留下来而 Secret 丢失，可以启动，但旧会话全部失效。

## CI 验证

[`.github/workflows/kind-postgres.yaml`](../../.github/workflows/kind-postgres.yaml) 会在 GitHub Actions 中：

- 运行 `helm lint`，渲染 SQLite、Ingress + 子路径、PostgreSQL 多副本和外部 Secret 场景，并断言不安全配置会被拒绝；
- 用仓库根目录的 `Dockerfile` 构建镜像并加载到 KinD；
- 在集群内启动 PostgreSQL，以 `database.mode=postgres` 安装 Chart；
- 通过端口转发访问 `/healthz`、`/api/status`，完成注册 → 登录 → 当前用户流程；
- 直接查询 PostgreSQL 的 `users` / `user_sessions` 表，确认数据写入的是外部数据库而不是 SQLite。

## 故障排查

| 现象 | 排查方向 |
|---|---|
| 登录返回 403 `CROSS_ORIGIN_REQUEST` | `PUBLIC_ORIGIN` 与浏览器实际地址（scheme + host + port）不一致 |
| 登录后立即掉线 | `SESSION_COOKIE_SECURE=true` 但通过 http 访问，浏览器丢弃 Cookie |
| Pod 一直 CrashLoop，日志提示数据库不可写 | PVC 权限或 `storageClassName`；容器用户需要对 `/data` 有写权限 |
| Pod 卡在 `wait-for-postgres` | DSN 错误、网络策略、`sslmode` 与证书不匹配 |
| 健康检查 404 | `config.BASE_PATH` 与探针路径不一致（应使用 Chart 自动生成的路径） |
| 升级后用户全部登出 | `SESSION_SECRET` 被替换（检查是否手工改过 Secret） |
