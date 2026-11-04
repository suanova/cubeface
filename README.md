# CubeFace

CubeFace 是一个面向 AI 模型的托管与分享平台（目标类似 Hugging Face / ModelScope）。当前是第一个可运行版本：**账号体系完整闭环**，首页内容暂留空，模型广场、模型上传与推理等能力尚未实现。

## 当前功能

- 用户名 + 密码注册、登录、登出、当前用户
- 邮箱验证码注册（6 位数字、10 分钟有效、最多 5 次尝试、60 秒重发冷却、按 IP 限流）
- OAuth 登录：GitHub、Google、通用 OIDC（可选启用）
- 数据库会话：随机不透明令牌放在 HttpOnly Cookie，数据库只存 HMAC 哈希；支持滑动续期、会话吊销、`auth_version` 全量失效
- 完整账号界面（登录、注册、OAuth、仪表盘）与中英双语、亮/暗主题
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
│   ├── service/             # 会话 / 用户 / 验证码 / 邮件 / OAuth
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
4. 通过 `/healthz`、`/api/status`、注册、登录、`/api/user/self` 验证可访问性；
5. 直接查询 PostgreSQL 主库确认账号与会话确实写入数据库；
6. 删除主库 Pod 触发自动故障切换，验证新主库选举成功且 CubeFace 恢复服务。

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

## API 概览

响应统一为：

```json
{ "success": true, "message": "", "data": {} }
{ "success": false, "code": "INVALID_CREDENTIALS", "message": "Invalid username or password" }
```

业务/校验错误使用 HTTP 200 + `success:false`；未登录 401、限流 429、跨源 403、未知错误 500。

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

## 测试

```bash
make test       # go test ./... + vitest
make check      # go vet + tsc + tests
```

- 后端集成测试使用临时 SQLite，覆盖注册/登录/登出、验证码、会话失效与滑动续期、同源校验、OAuth 错误跳转、子路径部署。
- 前端测试覆盖安全跳转、表单校验规则与品牌组件。
- PostgreSQL 可通过 `SQL_DSN=postgres://... go run .` 手动验证；自动化测试目前只使用 SQLite。

## 安全说明

- 密码使用 bcrypt；会话、验证码和 OAuth state 只存 HMAC，不下发明文。
- Cookie 为 `HttpOnly; SameSite=Lax`，生产默认 `Secure`；所有写请求额外校验 `Origin`/`Referer`。
- 登录、注册、验证码发送均有内存限流（单实例）。多实例部署时需替换为 Redis 等共享限流。
- 当前不提供找回密码、修改邮箱、2FA、Passkey、管理员后台。

## 许可证

本仓库以 [MIT](LICENSE) 许可证发布。

`helm/cubeface-chart/charts/cloudnative-pg` 是 [CloudNativePG](https://github.com/cloudnative-pg/charts) 的 vendored 副本，按其自带的 Apache-2.0 许可证分发（见该目录下的 `LICENSE`）。
