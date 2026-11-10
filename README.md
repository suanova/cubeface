# CubeFace

CubeFace 是一个面向 AI 模型的托管与分享平台（目标类似 Hugging Face / ModelScope）。当前是第一个可运行版本：**账号体系完整闭环**，并带有只读的模型库与数据集库（数据来自 CubeStack 集群 CRD）；模型上传、评测与在线推理尚未实现，首页内容暂留空。

## 当前功能

- 用户名 + 密码注册、登录、登出、当前用户
- 邮箱验证码注册（6 位数字、10 分钟有效、最多 5 次尝试、60 秒重发冷却、按 IP 限流）
- OAuth 登录：GitHub、Google、通用 OIDC（可选启用）
- 数据库会话：随机不透明令牌放在 HttpOnly Cookie，数据库只存 HMAC 哈希；支持滑动续期、会话吊销、`auth_version` 全量失效
- 完整账号界面（登录、注册、OAuth、仪表盘）与中英双语、亮/暗主题
- 模型库列表 `/models` 与详情 `/models/$name`：直接读取 CubeStack `ModelVersion` / `InferenceRuntimeProfile`，展示模型卡、存储摘要、兼容运行时与全部版本；集群不可用时优雅降级
- 数据集库列表 `/datasets` 与详情 `/datasets/$name`：读取 CubeStack `Dataset`，按语义名合并版本，展示格式、体积、记录数、版本与访问/去标识化/伦理等边界信息；与模型库独立降级
- 工作室 `/studio`：模型与数据集目录的入口页，展示两个目录的最近更新与接入状态；运行态、配额与审计尚未接入，页面明确标注而不展示估算数值
- 首页内容暂留空，仅保留站点导航
- 生产模式由 Go 二进制内嵌前端静态资源，并支持部署在子路径（`BASE_PATH`）

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.27、Gin、GORM、bcrypt、HMAC-SHA256、`net/smtp`、`golang.org/x/oauth2`、`coreos/go-oidc/v3` |
| 数据库 | SQLite（默认，纯 Go 驱动）或 PostgreSQL |
| 前端 | React 19、TypeScript、Rsbuild、TanStack Router / Query、Tailwind CSS v4、Zustand、react-hook-form + zod、i18next |
| 测试 | Go `testing` + `httptest`；Vitest + Testing Library |

## 目录结构

```
.
├── main.go                  # 服务入口、go:embed 前端产物
├── internal/
│   ├── config/              # 环境变量解析与校验
│   ├── common/              # 响应封装、加密、校验、内存限流
│   ├── model/               # GORM 模型、数据库初始化
│   ├── service/             # 会话 / 用户 / 验证码 / 邮件 / OAuth / 模型目录
│   ├── kube/                # 只读访问 Kubernetes API（CubeStack CRD）
│   ├── controller/          # HTTP 处理器
│   ├── middleware/          # 鉴权、同源校验、请求体限制
│   ├── router/              # 路由装配
│   └── webui/               # SPA 静态资源与 fallback
├── helm/cubeface-chart/     # Kubernetes Helm Chart
└── web/                     # Rsbuild 前端
```

## 快速开始（开发）

前置：Go 1.27+、[Bun](https://bun.sh/)。中国大陆网络可先配置 Go 代理。

```bash
cp .env.example .env
# 编辑 .env，至少确认 PUBLIC_ORIGIN 与前端实际地址一致（默认 http://localhost:5173）

# 首次运行需要先生成 web/dist（Go 通过 go:embed 内嵌它）
cd web && bun install && bun run build && cd ..

go run .
```

首次构建前端后，再启动前端开发服务器（热更新）：

```bash
cd web && bun run dev
```

打开 <http://localhost:5173>。开发模式下 `EMAIL_DEBUG_MODE=true`，注册验证码不会发邮件，而是打印在后端日志：

```
WARN [email-debug] verification email suppressed to=you@example.com purpose=register code=123456
```

> 后端默认监听 `:3000`，前端开发服务器监听 `:5173` 并把 `/api` 代理到后端。
> `PUBLIC_ORIGIN` 必须是**浏览器地址**（默认 `http://localhost:5173`），不是后端端口；否则 OAuth 状态 Cookie 与同源校验会失败。
> 如果 5173 被占用，Rsbuild 会改用其它端口，此时需同步修改 `PUBLIC_ORIGIN` 与 `PORT`/代理目标。

## 生产构建

```bash
make build          # 先构建 web/dist，再编译 bin/cubeface
```

运行前设置：

```bash
APP_ENV=production
PORT=3000
PUBLIC_ORIGIN=https://cubeface.example
SESSION_SECRET=$(openssl rand -base64 48)
SQL_DSN=postgres://user:pass@db:5432/cubeface?sslmode=disable   # 或留空使用 SQLite
EMAIL_VERIFICATION_ENABLED=true
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_TLS_MODE=starttls
SMTP_USERNAME=...
SMTP_PASSWORD=...
SMTP_FROM="CubeFace <no-reply@cubeface.example>"

./bin/cubeface
```

生产环境启动时会拒绝以下配置：缺少 `SESSION_SECRET` 或长度不足 32、调试邮箱模式未显式放行、开启邮箱验证但未配置 SMTP。生产默认 `SESSION_COOKIE_SECURE=true`，需要 HTTPS。

部署在子路径（例如 `https://example.com/cubeface/`）时：

```bash
BASE_PATH=/cubeface
PUBLIC_ORIGIN=https://example.com
```

Go 会在 SPA HTML 中注入 `<base href="/cubeface/">` 和 `window.__BASE_PATH__`，API、Cookie Path 与 OAuth 回调都会自动带上前缀。

## Kubernetes / Helm

`helm/cubeface-chart` 默认部署一套由 [CloudNativePG](https://cloudnative-pg.io/) 管理的高可用 PostgreSQL：Chart 内嵌 CloudNativePG operator 子 Chart，并渲染 `Cluster` CR（一主多从、流复制、自动故障切换、自签名 TLS）。CubeFace 本身仍是无状态单容器，可随 PostgreSQL 一起扩缩容。

```bash
helm install cubeface ./helm/cubeface-chart \
  --namespace cubeface --create-namespace

kubectl -n cubeface port-forward svc/cubeface 3000:80
# 打开 http://localhost:3000
```

默认创建 3 个 PostgreSQL 实例（`database.postgres.instances`）。也可以选择其它数据库后端：

```bash
# 外部 PostgreSQL（不安装 operator）
helm install cubeface ./helm/cubeface-chart -n cubeface --create-namespace \
  --set database.mode=external \
  --set database.postgres.operator.install=false \
  --set 'database.external.dsn=postgres://user:pass@db.example.com:5432/cubeface?sslmode=require'

# 单副本 SQLite（开发/小规模）
helm install cubeface ./helm/cubeface-chart -n cubeface --create-namespace \
  --set database.mode=sqlite \
  --set database.postgres.operator.install=false
```

Ingress、子路径、PostgreSQL 高可用/备份/PgBouncer、Secret 管理、邮箱与 OAuth 配置、升级注意事项见 [Chart README](helm/cubeface-chart/README.md)。Chart 默认使用的镜像由仓库根目录的 `Dockerfile` 构建（前端 + 后端多阶段构建，运行用户 UID 10001）。

CI 工作流 [`.github/workflows/kind-postgres.yaml`](.github/workflows/kind-postgres.yaml) 会：

1. `helm lint` 并渲染托管 PostgreSQL、SQLite、外部 DSN、外部 Secret、备份/PgBouncer 场景，拒绝不安全配置；
2. 构建镜像并加载到 KinD；
3. 以 `database.mode=postgres` 安装 Chart，由 Chart 自带的 CloudNativePG operator 创建 3 实例 PostgreSQL 集群；
4. 应用 `hack/kind` 下的最小 CubeStack CRD 与测试对象，验证 `/api/models` 与 `/api/datasets` 能读到 CR、合并同名版本、应用注解并脱敏存储信息；删除两个目录各自的 `ClusterRoleBinding` 后验证接口分别降级为 `FORBIDDEN`，恢复后重新可用；
5. 通过 `/healthz`、`/api/status`、注册、登录、`/api/user/self` 验证可访问性；
6. 直接查询 PostgreSQL 主库确认账号与会话确实写入数据库；
7. 删除主库 Pod 触发自动故障切换，验证新主库选举成功且 CubeFace 恢复服务。

镜像由两个工作流构建并推送到 `harbor.isuanova.com/suanova/cubeface`：

- [`.github/workflows/build-image-pr.yaml`](.github/workflows/build-image-pr.yaml)：PR 打开/更新时构建，推送 `pr-<number>` 与 `sha-<short>` 标签，用于预览环境；
- [`.github/workflows/build-image-push.yaml`](.github/workflows/build-image-push.yaml)：推送到 `main`/`master` 或 `v*` 标签时构建多架构镜像，推送分支名、`sha-<short>`、`v*` 对应的语义化版本，默认分支额外打 `latest`。

两个工作流都用 `vars.CI_BOT_NAME` / `secrets.CI_BOT_PASSWORD` 登录 Harbor。来自 fork 或 Dependabot 的 PR 拿不到 secret，此时只构建、不推送。Chart 默认镜像已指向该 Harbor 仓库（`image.repository`），Harbor 项目不公开时需要通过 `imagePullSecrets` 提供拉取凭据。

## 配置说明

完整列表见 [`.env.example`](.env.example)，常用项：

| 变量 | 默认值 | 说明 |
|---|---|---|
| `APP_ENV` | `development` | 设为 `production` 启用生产校验与安全默认值 |
| `PORT` | `3000` | Go 服务端口 |
| `BASE_PATH` | 空 | 子路径部署前缀 |
| `PUBLIC_ORIGIN` | `http://localhost:5173` | 浏览器访问源；同源校验与 OAuth 回调依赖它 |
| `SESSION_SECRET` | 开发固定值 | 生产必填，≥32 字符；轮换会让所有会话失效 |
| `REGISTER_ENABLED` | `true` | 关闭后禁止注册 |
| `SQL_DSN` | 空 | 以 `postgres://` / `postgresql://` 开头时使用 PostgreSQL |
| `SQLITE_PATH` | `data/cubeface.db` | `SQL_DSN` 为空时使用 |
| `EMAIL_VERIFICATION_ENABLED` | `true` | 注册是否要求邮箱验证码 |
| `EMAIL_DEBUG_MODE` | 非生产为 `true` | 不连 SMTP，验证码写入日志 |
| `SMTP_TLS_MODE` | `implicit` | `implicit`(465) / `starttls`(587) / `none` |
| `INITIAL_ADMIN_*` | 空 | 启动时按需创建 root 账号，已存在则跳过 |
| `MODEL_LIBRARY_ENABLED` | `true` | 模型库开关；关闭后接口返回 `available:false, reason:"DISABLED"` |
| `MODEL_LIBRARY_CACHE_TTL_SECONDS` | `30` | 模型目录快照缓存 TTL |
| `MODEL_LIBRARY_REQUEST_TIMEOUT_SECONDS` | `5` | 单次读取 Kubernetes API 的超时 |
| `MODEL_LIBRARY_API_URL` | 空 | 本地开发指向 `kubectl proxy`；留空则使用集群内 ServiceAccount |
| `DATASET_LIBRARY_ENABLED` | `true` | 数据集库开关；关闭后接口返回 `available:false, reason:"DISABLED"` |
| `DATASET_LIBRARY_CACHE_TTL_SECONDS` | `30` | 数据集目录快照缓存 TTL |
| `DATASET_LIBRARY_REQUEST_TIMEOUT_SECONDS` | `5` | 单次读取 Dataset CRD 的超时 |
| `DATASET_LIBRARY_API_URL` | 空 | 为空时回退 `MODEL_LIBRARY_API_URL` |

数据库在启动时通过 GORM `AutoMigrate` 自动建表；SQLite 默认开启 WAL 与 `busy_timeout`，并将连接数限制为 1（仅适合单实例）。

### OAuth 配置

只需填写想启用的提供商，未配置的不会出现在登录页。

| 提供商 | 变量 | 回调 URL |
|---|---|---|
| GitHub | `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | `${PUBLIC_ORIGIN}${BASE_PATH}/api/oauth/github/callback` |
| Google | `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | `${PUBLIC_ORIGIN}${BASE_PATH}/api/oauth/google/callback` |
| 通用 OIDC | `OIDC_ISSUER_URL` / `OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` | `${PUBLIC_ORIGIN}${BASE_PATH}/api/oauth/oidc/callback` |

- Google 使用 OIDC discovery，会校验 `iss`、`aud`、`exp`、`nonce`。
- 通用 OIDC 的显示名称由 `OIDC_DISPLAY_NAME` 控制。
- 若回调地址与 `PUBLIC_ORIGIN` 不同（例如独立代理），可设置 `OAUTH_REDIRECT_BASE`。
- 首次 OAuth 登录会创建账号；若提供商返回已验证邮箱且本地已有同邮箱账号，会自动绑定。

## 模型库（CubeStack）

`/models` 与 `/models/$name` 不读取 CubeFace 数据库，而是实时读取集群里的 CubeStack 自定义资源（`ai.cubestack.io/v1alpha1`）：

- `ModelVersion`：按 `spec.model` 分组，同一模型的多个版本合并为一张卡；`spec.version` 按自然序降序，最新版本置顶。
- `InferenceRuntimeProfile`：仅当 profile 的 `spec.modelRequirements.architectures` **与** `spec.modelRequirements.quantization` 同时包含版本的值时才视为兼容（空列表不是通配）。

页面另需的展示信息来自注解（同一模型取最新版本的注解）：

| 注解 | 用途 |
|---|---|
| `ai.cubestack.io/description` | 模型/运行时简介 |
| `cubeface.io/display-name` | 展示名，回退 `spec.model` |
| `cubeface.io/publisher` | 发布者 |
| `cubeface.io/task` | 任务类型（筛选项） |
| `cubeface.io/license` | 许可证 |
| `cubeface.io/tags` | 逗号分隔标签（最多 12 个） |
| `cubeface.io/usage-boundary` | 使用边界提示 |
| `cubeface.io/featured` | `"true"` 时作为精选 |

接口不会下发 `storage.hostPath.path`、`storage.s3.uri`、`credentialsRef.name`、`status.rootPath`，只返回存储策略、容量以及“是否已配置”的布尔值。

集群不可用、CRD 未安装、缺少 RBAC 时，列表与详情仍返回 HTTP 200，带 `available:false` 和原因码（`DISABLED` / `NO_CLUSTER` / `CRD_MISSING` / `FORBIDDEN` / `UNAUTHORIZED` / `TIMEOUT` / `API_ERROR`），前端显示可重试的降级状态；只有集群健康但模型确实不存在时，详情才返回 404 `MODEL_NOT_FOUND`。

本地开发用 `kubectl proxy` 提供 API（无需 kubeconfig 解析）：

```bash
kubectl proxy --port=8001
MODEL_LIBRARY_API_URL=http://127.0.0.1:8001 go run .
curl -s localhost:3000/api/models | jq '.data.available, .data.total'
```

Helm Chart 默认创建 ServiceAccount（`serviceAccount.create=true`）和一个只读 ClusterRole/ClusterRoleBinding（`modelLibrary.rbac.create=true`），仅授予 `ai.cubestack.io` 下 `modelversions`、`inferenceruntimeprofiles` 的 `list` 权限；CRD 是集群级资源，因此必须用 ClusterRole。不需要模型库时用 `--set modelLibrary.enabled=false` 关闭页面数据与 RBAC；自行管理权限时用 `--set modelLibrary.rbac.create=false`。

## 数据集库（CubeStack）

`/datasets` 与 `/datasets/$name` 读取集群里的 `datasets.ai.cubestack.io/v1alpha1`：

- `Dataset`：按 `spec.dataset` 分组，同一数据集的多个 `spec.version` 合并为一条记录；版本按自然序降序。
- 结构化字段来自 spec：`format`、`size`、`recordCount`、`storage`；展示字段来自注解（取最新版本）：

| 注解 | 用途 |
|---|---|
| `ai.cubestack.io/description` | 数据集简介 |
| `cubeface.io/display-name` | 展示名，回退 `spec.dataset` |
| `cubeface.io/publisher` | 发布者 |
| `cubeface.io/task` | 任务类型（筛选项） |
| `cubeface.io/license` | 许可证 |
| `cubeface.io/tags` | 逗号分隔标签 |
| `cubeface.io/access` | 获取方式（如 public / gated） |
| `cubeface.io/deidentification` | 去标识化说明 |
| `cubeface.io/ethics` | 伦理依据 |
| `cubeface.io/source-repo` | 来源仓库 |
| `cubeface.io/usage-boundary` | 使用边界提示 |
| `cubeface.io/featured` | `"true"` 时作为精选 |

存储脱敏规则与模型库一致：不下发 `storage.hostPath.path`、`storage.s3.uri`、`credentialsRef.name`。数据集目录有独立的缓存与开关（`DATASET_LIBRARY_*`），datasets CRD 缺失不会影响 `/api/models`，反之亦然；不可用时返回 200 + `available:false`，健康但名称不存在时返回 404 `DATASET_NOT_FOUND`。

```bash
kubectl proxy --port=8001
DATASET_LIBRARY_API_URL=http://127.0.0.1:8001 go run .
curl -s localhost:3000/api/datasets | jq '.data.available, .data.total'
```

Helm Chart 为数据集库单独创建只读 ClusterRole/ClusterRoleBinding，仅授予 `datasets` 的 `list` 权限（`datasetLibrary.*`，与模型库互不影响）。

## API 概览

响应统一为：

```json
{ "success": true, "message": "", "data": {} }
{ "success": false, "code": "INVALID_CREDENTIALS", "message": "Invalid username or password" }
```

业务/校验错误通常使用 HTTP 200 + `success:false`；未登录 401、限流 429、跨源 403、未知错误 500。模型详情在集群健康但模型不存在时返回 404 + `MODEL_NOT_FOUND`。

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/status` | 系统信息、注册/邮箱验证开关、已启用 OAuth 提供商 |
| `GET` | `/healthz` | 数据库连通性检查 |
| `POST` | `/api/user/register` | 注册 `{username,email,password,verification_code}` |
| `POST` | `/api/user/login` | 登录 `{username,password}`，`username` 也可传邮箱 |
| `POST` | `/api/user/auth/logout` | 登出（幂等） |
| `GET` | `/api/user/self` | 当前用户（需登录） |
| `POST` | `/api/verification/send` | 发送注册验证码 `{email}` |
| `GET` | `/api/oauth/:provider/start` | 跳转到 OAuth 提供商 |
| `GET` | `/api/oauth/:provider/callback` | OAuth 回调 |
| `GET` | `/api/models` | 模型库列表（公开；集群异常时 200 + `available:false`） |
| `GET` | `/api/models/:name` | 模型详情（`spec.model` 或 `ModelVersion.metadata.name`；不存在 404） |
| `GET` | `/api/datasets` | 数据集库列表（公开；集群异常时 200 + `available:false`） |
| `GET` | `/api/datasets/:name` | 数据集详情（`spec.dataset` 或 `Dataset.metadata.name`；不存在 404 `DATASET_NOT_FOUND`） |

## 测试

```bash
make test       # go test ./... + vitest
make check      # go vet + tsc + tests
```

- 后端集成测试使用临时 SQLite，覆盖注册/登录/登出、验证码、会话失效与滑动续期、同源校验、OAuth 错误跳转、子路径部署。
- 模型目录测试覆盖版本分组与自然排序、注解回退、兼容规则、存储脱敏、TTL/负缓存/stale 降级与 CRD 错误分类。
- 数据集目录测试覆盖语义分组、版本排序、注解回退、存储脱敏，以及独立缓存与降级路径。
- 前端测试覆盖安全跳转、表单校验规则、品牌组件，以及模型/数据集的筛选、排序与搜索参数解析。
- PostgreSQL 可通过 `SQL_DSN=postgres://... go run .` 手动验证；自动化测试目前只使用 SQLite。

## 安全说明

- 密码使用 bcrypt；会话、验证码和 OAuth state 只存 HMAC，不下发明文。
- Cookie 为 `HttpOnly; SameSite=Lax`，生产默认 `Secure`；所有写请求额外校验 `Origin`/`Referer`。
- 登录、注册、验证码发送均有内存限流（单实例）。多实例部署时需替换为 Redis 等共享限流。
- 当前不提供找回密码、修改邮箱、2FA、Passkey、管理员后台。

## 许可证

本仓库以 [MIT](LICENSE) 许可证发布。

`helm/cubeface-chart/charts/cloudnative-pg` 是 [CloudNativePG](https://github.com/cloudnative-pg/charts) 的 vendored 副本，按其自带的 Apache-2.0 许可证分发（见该目录下的 `LICENSE`）。
