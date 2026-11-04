# CubeFace Helm Chart

把 CubeFace 部署到 Kubernetes。CubeFace 是单个 Go 二进制（前端已内嵌），因此本 Chart 只有一个无状态 Deployment 和 Service，不需要 Redis。

数据库有三种模式：

- **`postgres`（默认）**：Chart 内嵌 [CloudNativePG](https://cloudnative-pg.io/) operator 子 Chart，并创建 `Cluster` CR。默认 3 个 PostgreSQL 实例，流复制、自动故障切换、自签名 TLS，可水平扩容 CubeFace。
- **`external`**：连接已有的 PostgreSQL，DSN 由使用者提供。
- **`sqlite`**：数据库文件放在 PVC 的 `/data/cubeface.db`，只能单副本运行，适合开发或小规模部署。

## 前置条件

- Kubernetes 1.29+（CloudNativePG 1.30 子 Chart 的要求）
- Helm 3.8+
- 可用的 CubeFace 镜像。默认 `ghcr.io/cubeface/cubeface:<appVersion>`；也可以自行构建并推送到私有仓库后覆盖 `image.repository` / `image.tag`。
- `postgres` / `sqlite` 模式需要一个默认 StorageClass，或在安装时显式指定。
- 安装 CloudNativePG operator 需要 cluster 级权限（CRD、ClusterRole、Webhook）。若集群已安装 operator，设置 `database.postgres.operator.install=false`。

## 快速安装（托管 PostgreSQL，默认）

```bash
helm install cubeface ./helm/cubeface-chart \
  --namespace cubeface --create-namespace
```

这会安装：

- CloudNativePG operator（`charts/cloudnative-pg`，operator 1.30.0）；
- 名为 `<release>-postgres` 的 3 实例 PostgreSQL 集群；
- CubeFace Deployment，通过 `<release>-postgres-rw:5432` 连接主库（故障切换后 Service 自动跟随新主库）。

默认没有 Ingress，使用端口转发访问。本地端口要与 `config.PORT` 一致，否则浏览器 Origin 和 `PUBLIC_ORIGIN` 不匹配：

```bash
kubectl -n cubeface port-forward svc/cubeface 3000:80
# 打开 http://localhost:3000
```

查看数据库状态：

```bash
kubectl -n cubeface get cluster cubeface-postgres
kubectl -n cubeface get pods -l cnpg.io/cluster=cubeface-postgres
```

首次安装会生成 `SESSION_SECRET` 并写入 `cubeface-secret`，生成 PostgreSQL 应用密码并写入 `cubeface-postgres-app-auth` 与 `cubeface-postgres-dsn`。请备份它们。

## PostgreSQL 高可用

### 实例数与资源

```yaml
database:
  mode: postgres
  postgres:
    instances: 3          # >= 2 才有 HA；3 是推荐值
    storage: 10Gi
    storageClassName: ""  # 空表示使用默认 StorageClass
    resources:
      requests: {cpu: 500m, memory: 1Gi}
      limits: {cpu: "2", memory: 4Gi}
    auth:
      database: cubeface
      username: cubeface  # 只能是 [a-z_][a-z0-9_]*
```

`instances=1` 可以运行，但没有故障切换能力。生产环境建议 3 个实例并配合 `database.postgres.storageClassName` 选择支持跨节点访问的存储（`ReadWriteOnce` 即可，CNPG 每个实例使用独立 PVC）。

CubeFace 的 `replicaCount` 与 PostgreSQL 实例数相互独立。多个应用副本会共享同一个 `-rw` 主库 Service；应用启动时执行 `AutoMigrate`，多副本同时启动通常安全，但大版本升级时建议先把 `replicaCount` 缩到 1。

### 自动故障切换

CloudNativePG 自动检测主库故障并提升一个副本。应用只需连接 `-rw` Service，不需要知道当前主库是谁。等待初始化容器会持续重试 `psql "SELECT 1"`，直到新主库可写。

### 备份

```yaml
database:
  postgres:
    backups:
      enabled: true
      schedule: "0 0 2 * * 0"
      method: volumeSnapshot
```

- `volumeSnapshot`：需要 VolumeSnapshot CRD 和 VolumeSnapshotClass。
- `fileArchiving`：需要配置 `barmanObjectStore`（S3/兼容对象存储）。

### PgBouncer

```yaml
database:
  postgres:
    pgBouncer:
      enabled: true
      replicas: 2
```

会额外创建 `Pooler`（`<cluster>-pgbouncer:5432`）。CubeFace 自身仍直连 `-rw`；该池化服务供其它客户端使用。

### 复用已有 CloudNativePG operator

```bash
--set database.postgres.operator.install=false
```

此时 Chart 不安装 operator 子 Chart，但仍渲染 `Cluster` CR，由集群中已有的 operator 接管。若集群没有 operator，安装会因找不到 `postgresql.cnpg.io/v1` CRD 而失败。

### CRD 升级

CloudNativePG 的 CRD 位于子 Chart 的 `crds/` 目录。Helm 只在首次安装时创建 CRD，升级或卸载不会更新/删除它们。升级 operator 时如果 CRD 有变更，需要先手动应用新版 `charts/cloudnative-pg/crds/crds.yaml`。

## 使用 SQLite（开发/小规模）

```bash
helm install cubeface ./helm/cubeface-chart \
  --namespace cubeface --create-namespace \
  --set database.mode=sqlite \
  --set database.postgres.operator.install=false
```

- 数据库文件在 PVC 的 `/data/cubeface.db`；`persistence.retain=true`（默认）会加 `helm.sh/resource-policy: keep`，卸载后保留数据。
- SQLite 是单写者，`replicaCount` 必须为 1；Chart 会拒绝多副本。
- 使用 `Recreate` 更新策略，升级期间会短暂中断。
- 即使使用 SQLite，默认的 `database.postgres.operator.install=true` 仍会安装 CloudNativePG operator；用不到时请显式设为 `false`（如上）。

## 使用外部 PostgreSQL

```bash
helm upgrade --install cubeface ./helm/cubeface-chart \
  --namespace cubeface \
  --set database.mode=external \
  --set database.postgres.operator.install=false \
  --set 'database.external.dsn=postgres://cubeface:password@db.example.com:5432/cubeface?sslmode=require' \
  --set replicaCount=2
```

- DSN 只写入 Secret，不会进入 ConfigMap。
- `database.external.waitForPostgres=true`（默认）会先运行 `psql -tAc "SELECT 1"`，直到数据库可连接。
- 此模式不渲染 PVC，`persistence.*` 被忽略。
- 多副本时自动渲染 PDB，并给 Pod 添加软反亲和（可用 `podAntiAffinity=false` 关闭）。
- 应用启动时会执行 `AutoMigrate`；DSN 用户需要在其 schema 上拥有建表/改表权限。

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

托管 PostgreSQL 模式还会创建以下两个 Chart 管理的 Secret（无论 `secret.create` 是否为 true）：

| Secret | 内容 | 说明 |
|---|---|---|
| `<cluster>-app-auth` | `username` / `password` | CloudNativePG `bootstrap.initdb.secret`；operator 用它设置数据库 owner 密码 |
| `<release>-postgres-dsn` | `SQL_DSN` | 应用实际使用的连接串，指向 `<cluster>-rw`，`sslmode=require` |

PostgreSQL owner 密码来自 `secrets.POSTGRES_PASSWORD`；为空时生成 24 位随机值，并通过 `lookup` 在升级时复用。密码不能包含空格或 URI 特殊字符（`@ : / ? # [ ] %`），否则 Chart 在渲染阶段报错。

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
| `SQL_DSN` | `database.mode=external` |
| `SMTP_PASSWORD` | 使用 SMTP 时 |
| `GITHUB_CLIENT_SECRET` / `GOOGLE_CLIENT_SECRET` / `OIDC_CLIENT_SECRET` | 对应 OAuth 提供商 |
| `INITIAL_ADMIN_PASSWORD` | 使用初始化管理员时 |

其余 key 可省略。Deployment 对可选 key 使用 `optional: true`。托管 PostgreSQL 模式下 `SQL_DSN` 由 Chart 自己生成，不需要放进外部 Secret。

## 常用配置

| 值 | 默认 | 说明 |
|---|---|---|
| `image.repository` / `image.tag` | `ghcr.io/cubeface/cubeface` / Chart appVersion | 镜像地址 |
| `replicaCount` | `1` | SQLite 模式必须为 1 |
| `database.mode` | `postgres` | `postgres` / `external` / `sqlite` |
| `database.postgres.operator.install` | `true` | 是否安装内嵌的 CloudNativePG operator |
| `database.postgres.instances` | `3` | PostgreSQL 实例数（>=2 才有 HA） |
| `database.postgres.storage` | `10Gi` | 每个实例的 PVC 容量 |
| `database.postgres.image` | `ghcr.io/cloudnative-pg/postgresql:16.14-system-trixie` | PostgreSQL 镜像 |
| `database.postgres.retain` | `true` | Cluster 加 `helm.sh/resource-policy: keep` |
| `database.external.dsn` | 空 | 外部 PostgreSQL DSN（写入 Secret） |
| `secrets.POSTGRES_PASSWORD` | 空（自动生成） | 托管 PostgreSQL owner 密码 |
| `persistence.enabled` | `true` | SQLite 数据卷；关闭则使用 `emptyDir`（仅供测试） |
| `persistence.size` | `1Gi` | SQLite PVC 容量 |
| `persistence.retain` | `true` | SQLite PVC 卸载后保留 |
| `config.BASE_PATH` | 空 | 子路径前缀 |
| `config.PUBLIC_ORIGIN` | 自动 | 浏览器访问源，必须与真实地址一致 |
| `config.SESSION_COOKIE_SECURE` | 自动 | 空=按 `PUBLIC_ORIGIN` 是否 https 决定 |
| `config.TRUSTED_PROXIES` | 空 | 反向代理 CIDR，逗号分隔，用于取得真实客户端 IP |
| `config.REGISTER_ENABLED` | `"true"` | 是否开放注册 |
| `config.INITIAL_ADMIN_USERNAME` | 空 | 首次启动创建 root 账号的用户名 |
| `ingress.enabled` | `false` | 是否创建 Ingress |
| `service.port` | `80` | Service 端口 |
| `pdb.enabled` / `minAvailable` | `true` / `1` | 仅在应用副本数 > 1 时渲染 |
| `resources` | 100m/128Mi → 1/512Mi | CubeFace 容器资源 |
| `cloudnative-pg.image.tag` | `1.30.0` | 内嵌 operator 镜像版本 |
| `envVars` | `[]` | 追加环境变量，例如 `[{name: DB_MAX_OPEN_CONNS, value: "20"}]` |
| `extraVolumes` / `extraVolumeMounts` | `[]` | 追加卷和挂载 |

完整列表和注释见 [`values.yaml`](values.yaml)。

## 健康检查

探针请求 `${BASE_PATH}/healthz`，该接口会执行数据库 `SELECT 1`。启动探针允许最多 5 分钟（`failureThreshold: 60 × 5s`）；托管 PostgreSQL 首次 bootstrap 的时间由 `wait-for-postgres` 初始化容器承担，应用容器在数据库可写后才启动。

## 校验规则

Chart 在渲染阶段会拒绝以下配置：

- `database.mode` 不是 `postgres` / `external` / `sqlite`
- SQLite 模式下 `replicaCount > 1`
- 外部 PostgreSQL 模式 + `secret.create=true` 但 DSN 为空
- 托管 PostgreSQL 的 `auth.username` / `auth.database` 不是合法 PostgreSQL 标识符
- 托管 PostgreSQL 的 `instances < 1`、`storage` 为空
- `secrets.POSTGRES_PASSWORD` 含 URI 特殊字符
- 开启备份但 `method` 非法，或 `fileArchiving` 缺少 `barmanObjectStore`
- 开启 PgBouncer 但 `replicas < 1`
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
- `database.postgres.retain=true` 时 `helm uninstall` 不会删除 CloudNativePG `Cluster` 及其 PVC。重新安装前如果需要彻底清理：

  ```bash
  kubectl -n cubeface delete cluster cubeface-postgres
  ```

- `persistence.retain=true` 时 SQLite PVC 同样会保留；需要清理时手动删除：

  ```bash
  kubectl -n cubeface delete pvc cubeface-data
  ```

- 自动生成的 `SESSION_SECRET` 在卸载后随 Secret 一起删除；如果数据库保留下来而 Secret 丢失，可以启动，但旧会话全部失效。

## CI 验证

[`.github/workflows/kind-postgres.yaml`](../../.github/workflows/kind-postgres.yaml) 会在 GitHub Actions 中：

- 运行 `helm lint`，渲染托管 PostgreSQL、SQLite、外部 DSN、外部 Secret、备份/PgBouncer 场景，并断言不安全配置会被拒绝；
- 用仓库根目录的 `Dockerfile` 构建镜像并加载到 KinD；
- 以默认的 `database.mode=postgres` 安装 Chart，由 Chart 自带的 CloudNativePG operator 创建 3 实例 PostgreSQL 集群；
- 通过端口转发访问 `/healthz`、`/api/status`，完成注册 → 登录 → 当前用户流程；
- 直接查询 PostgreSQL 主库的 `users` / `user_sessions` 表，确认数据写入的是 CNPG 集群而不是 SQLite；
- 删除主库 Pod 触发自动故障切换，验证新主库产生、集群恢复到 3 个 ready 实例且 CubeFace 继续服务。

## 故障排查

| 现象 | 排查方向 |
|---|---|
| 登录返回 403 `CROSS_ORIGIN_REQUEST` | `PUBLIC_ORIGIN` 与浏览器实际地址（scheme + host + port）不一致 |
| 登录后立即掉线 | `SESSION_COOKIE_SECURE=true` 但通过 http 访问，浏览器丢弃 Cookie |
| Pod 一直 CrashLoop，日志提示数据库不可写 | SQLite 模式的 PVC 权限或 `storageClassName`；容器用户需要对 `/data` 有写权限 |
| Pod 卡在 `wait-for-postgres` | DSN 错误、网络策略、`sslmode` 与证书不匹配；托管模式检查 Cluster 是否 Ready |
| `Cluster` 一直不 Ready | `kubectl -n <ns> describe cluster <name>`、查看 `cnpg.io/cluster` Pod 日志、StorageClass 是否可动态供给 |
| 安装报 `no matches for kind "Cluster"` | 未安装 CRD：保留 `database.postgres.operator.install=true`，或先安装 CloudNativePG operator |
| 健康检查 404 | `config.BASE_PATH` 与探针路径不一致（应使用 Chart 自动生成的路径） |
| 升级后用户全部登出 | `SESSION_SECRET` 被替换（检查是否手工改过 Secret） |
