# OmniCred 本地账号密码管理器设计文档

## 1. 文档目的

本文档定义 OmniCred 第一版的架构、数据模型、HTTP API、目录结构、实现步骤和验收标准。实现人员应能根据本文档完成一个可运行、可测试的本地账号密码管理服务。

## 2. 项目目标

OmniCred 用于在本机保存和管理账号信息，第一版需要支持：

- 保存 GitHub、Google、ChatGPT、OmniCred 等平台的账号信息。
- 双击 `OmniCred.exe` 后通过独立桌面窗口管理账号信息，不打开浏览器。
- 通过本地 HTTP API 新增、读取、修改和删除账号信息。
- 使用 Go 编写。
- 使用本地文件型数据库持久化数据。
- 保持代码结构简单，并允许以后替换存储方式或增加功能。

每条账号记录包含以下核心信息：

- `provider`：平台名称，例如 `github`、`google`、`chatgpt`、`omnicred`。
- `account`：实际登录账号，例如邮箱、手机号或登录 ID。
- `username`：用户名或昵称；没有时可以为空。
- `password`：登录密码。

## 3. 明确边界

### 3.1 第一版必须实现

- 本地 SQLite 持久化。
- Wails 独立桌面窗口和单实例运行。
- React + TypeScript 单页管理界面。
- 使用 shadcn/ui 构建可访问的现代界面，并支持明暗主题和响应式布局。
- 完整的账号 CRUD API。
- 基本字段校验。
- API 版本前缀 `/api/v1`。
- 健康检查接口。
- 数据库自动初始化和版本迁移。
- 默认只监听 `127.0.0.1`。
- 日志不得输出账号密码和请求体。
- 单元测试与 HTTP 集成测试。
- 前端组件测试和关键交互测试。
- 所有手写代码文件不超过 400 行。
- Windows 10/11 amd64 上可构建和运行桌面 EXE。

### 3.2 第一版不实现

- 密码加密。
- 主密码或 API 登录认证。
- 浏览器插件。
- 多用户和权限系统。
- 云同步、多设备同步或远程访问。
- 自动登录第三方网站。
- 密码生成器、密码强度检查和历史版本。

这些功能以后可以增加，但不得为了未来功能提前引入插件系统、微服务或复杂配置框架。

## 4. 安全声明

第一版按照需求不加密数据。SQLite 文件不是加密容器，能够读取数据库文件的用户或程序可以提取其中的明文密码。

即使不加密，也必须遵守以下最低安全要求：

1. 服务只监听 `127.0.0.1`，不得监听 `0.0.0.0`。
2. 数据库文件不得提交到 Git。
3. 不记录请求体、响应体、账号或密码。
4. HTTP 写接口只接受 `application/json`。
5. CORS 只允许 Wails 内部来源 `http://wails.localhost` 和 `https://wails.localhost`。
6. 启动日志可以记录数据库路径，但不得记录数据库内容。
7. 数据库文件应尽可能限制为当前系统用户可读写。
8. 前端不得把密码写入 `localStorage`、`sessionStorage` 或 IndexedDB。
9. 前端不得在 WebView 开发者控制台打印 API 响应或账号数据。
10. 前端生产构建嵌入桌面 EXE，不依赖公网 CDN、在线字体或远程脚本。

这些措施只能降低意外暴露风险，不能防止本机恶意程序访问数据。

## 5. 总体架构

OmniCred 使用 Wails 承载 React 单页应用，并在同一个 Go 进程中运行 SQLite 和本地 REST API：

```text
                    OmniCred.exe
                         │
             ┌───────────┴───────────┐
             ▼                       ▼
     Wails 独立桌面窗口       127.0.0.1:8787
     React + 系统 WebView       REST API
             │                       │
             └───────────┬───────────┘
                         ▼
               Credential Service
               字段校验、业务规则
                         │
                         ▼
                     Store 接口
                         │
                         ▼
                   SQLite Store
                         │
                         ▼
          %APPDATA%/OmniCred/omnicred.db
```

调用方向必须保持单向：

```text
React -> HTTP API -> credential service -> credential store interface <- sqlite
```

React 通过固定的本地 HTTP API 访问数据；本地脚本也可以调用同一 API。HTTP 层不得直接执行 SQL，SQLite 层不得依赖 HTTP、Wails 或 React 类型。

## 6. 技术选择

### 6.1 桌面容器

使用稳定版 Wails v2 将 React、字体和图标嵌入 Windows 桌面 EXE。桌面程序使用系统 WebView2 Runtime 渲染界面，不启动默认浏览器，也不需要用户单独运行 Node.js。

桌面窗口要求：

- 默认大小 `1280 × 820`，最小大小 `760 × 560`。
- 关闭主窗口即退出程序并关闭本地 API。
- 使用单实例锁；再次启动时唤醒已有窗口。
- 生产模式关闭 WebView 默认右键菜单和开发者工具。
- 应用图标、前端静态资源和 Go 后端一起构建进 EXE。

### 6.2 HTTP 服务

使用 Go 标准库 `net/http`，第一版不引入 Gin、Echo 等 Web 框架。当前接口数量很少，标准库足以完成路由、JSON 处理、中间件和测试。

### 6.3 数据存储

使用 SQLite，理由如下：

- 数据存储在单个本地文件中。
- 支持事务和可靠写入。
- 查询和后续迁移比手写 JSON 文件简单。
- 不需要安装独立数据库服务。

实现时选择一个纯 Go SQLite 驱动并固定依赖版本，避免要求用户安装 C 编译器。SQLite 驱动只能出现在 `internal/sqlite` 包和程序装配入口附近，业务层不得依赖具体驱动。

### 6.4 日志

使用 Go 标准库日志能力。每次请求最多记录：

- HTTP 方法。
- URL 路径。
- HTTP 状态码。
- 请求耗时。

禁止记录请求体、响应体和 `Credential` 的完整结构。

### 6.5 前端技术栈

前端采用以下现代技术栈：

- React 19 + TypeScript。
- Vite 8，用于开发服务器和生产构建。
- Tailwind CSS 4，用于设计 token 和布局样式。
- shadcn/ui + Radix UI，用于 Button、Dialog、Table、Form 等基础组件。
- Lucide React，作为全站唯一图标体系。
- TanStack Query，管理 API 查询、缓存和写入后的失效刷新。
- React Hook Form + Zod，管理表单状态和客户端校验。
- Vitest + Testing Library，完成前端测试。

依赖安装时固定实际版本到 lockfile。只添加页面实际使用的 shadcn/ui 组件，不批量引入整个组件集合。

第一版是单页面账号管理工具，不引入 Next.js、服务端渲染或全局状态框架。除非以后出现多个可导航页面，否则不引入 React Router。

### 6.6 前端视觉方向

界面采用现代、克制的本地密码保险库风格：

- 深蓝黑色作为主要背景，绿色作为主要操作强调色。
- 使用清晰的卡片或响应式表格展示账号，不使用传统拥挤后台样式。
- 桌面端优先展示高效管理布局，最小支持 375px 宽度。
- 支持亮色和暗色主题，两套主题分别检查对比度。
- 密码默认显示为掩码，通过按钮临时显示或复制。
- 动画只用于 Dialog、Toast、Hover 和状态变化，并遵守 `prefers-reduced-motion`。
- 所有表单都有可见 Label，图标按钮必须具有可访问名称和可见焦点状态。

首屏必须提供：

1. OmniCred 标识和“仅本地运行”状态。
2. 平台筛选和关键词搜索。
3. 新增账号主按钮。
4. 账号卡片或表格列表。
5. 查看、复制、编辑和删除操作。
6. 加载、空数据、保存成功和请求失败状态。

### 6.7 代码文件行数限制

所有手写代码文件的物理总行数不得超过 400 行。总行数包括代码、注释和空行，这样检查规则保持客观一致。

限制适用于：

- Go：`*.go`
- TypeScript/JavaScript：`*.ts`、`*.tsx`、`*.js`、`*.jsx`
- 样式：`*.css`
- 数据库脚本：`*.sql`

以下内容不参与检查：

- `node_modules/` 和其他第三方依赖目录。
- `dist/` 等生产构建产物。
- `go.sum`、`package-lock.json` 等依赖锁文件。
- 明确标记且不由项目维护者手工修改的工具生成文件。

shadcn/ui 添加到仓库中的组件属于项目源代码，仍然受 400 行限制。文件达到约 350 行时应主动按职责拆分，避免在最后才进行机械拆分。禁止通过压缩代码、合并多条语句或删除有价值的可读性来满足行数限制。

仓库应提供跨平台检查工具 `tools/linecheck`：

```text
go run ./tools/linecheck
```

该命令发现任何受管文件超过 400 行时必须返回非零退出码，并在测试或 CI 验收中执行。

## 7. 数据模型

领域模型：

```go
package credential

import "time"

type Credential struct {
    ID        int64     `json:"id"`
    Provider  string    `json:"provider"`
    Account   string    `json:"account"`
    Username  string    `json:"username"`
    Password  string    `json:"password"`
    TOTPSecret string    `json:"totp_secret"`
    RecoveryCodes []string `json:"recovery_codes"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Filter struct {
    Provider string
    Query    string
}
```

字段规则：

| 字段 | 必填 | 规则 |
|---|---:|---|
| `provider` | 是 | 去除首尾空格后不能为空；保存为小写；最大 100 个字符 |
| `account` | 是 | 去除首尾空格后不能为空；最大 4096 个字符 |
| `username` | 否 | 去除首尾空格；最大 4096 个字符 |
| `password` | 是 | 不能为空；不去除空格；最大 16384 个字符 |
| `totp_secret` | 否 | Base32 TOTP 密钥；去除空格和横线并转为大写；最大 512 个字符 |
| `recovery_codes` | 否 | 一次性恢复码列表；最多 100 个；每个最多 256 个字符；不允许重复 |

密码不能执行 `TrimSpace`，因为空格可能是密码的有效组成部分。

第一版不为 `provider` 定义枚举。增加新平台只需写入新的平台名称，不需要修改代码和数据库结构。

## 8. 数据库设计

初始 migration：

```sql
CREATE TABLE IF NOT EXISTS credentials (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    provider   TEXT NOT NULL,
    account    TEXT NOT NULL,
    username   TEXT NOT NULL DEFAULT '',
    password   TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_credentials_provider
    ON credentials(provider);

PRAGMA user_version = 1;
```

`003_add_totp_secret.sql` 为现有数据库增加 `totp_secret TEXT NOT NULL DEFAULT ''`，并将 `user_version` 更新为 `3`。

`004_add_recovery_codes.sql` 增加 `recovery_codes TEXT NOT NULL DEFAULT '[]'`，以 JSON 数组保存恢复码，并将 `user_version` 更新为 `4`。

时间统一保存为 UTC 的 RFC 3339 格式，API 也返回 RFC 3339 字符串。

第一版不增加 `(provider, account)` 唯一约束，因为同一平台可能存在相同登录账号但用途不同的记录。是否禁止重复应在有明确产品规则后再决定。

### 8.1 数据库初始化

启动时按以下顺序执行：

1. 创建数据库所在目录。
2. 打开 SQLite 数据库。
3. 执行 `PingContext`。
4. 读取 `PRAGMA user_version`。
5. 在事务中依次执行尚未应用的 migration。
6. migration 成功后再启动 HTTP 服务。

任何 migration 失败都必须终止启动，不能让服务运行在未知表结构上。

第一版可以把 SQL migration 通过 `//go:embed` 编译进可执行文件，避免运行时依赖工作目录中的 SQL 文件。

## 9. Store 接口

只在持久化边界定义接口：

```go
package credential

import (
    "context"
    "errors"
)

var ErrNotFound = errors.New("credential not found")

type Store interface {
    Create(ctx context.Context, item Credential) (Credential, error)
    Get(ctx context.Context, id int64) (Credential, error)
    List(ctx context.Context, filter Filter) ([]Credential, error)
    Update(ctx context.Context, item Credential) (Credential, error)
    Delete(ctx context.Context, id int64) error
}
```

Store 实现必须：

- 所有数据库调用都接收 `context.Context`。
- 使用参数化 SQL，不拼接用户输入。
- 将 `sql.ErrNoRows` 转换为 `credential.ErrNotFound`。
- 更新和删除不存在的 ID 时返回 `credential.ErrNotFound`。
- `List` 没有结果时返回空数组，而不是 `null`。

## 10. Service 设计

Service 负责字段校验、格式规范化和时间生成，不执行 SQL。

建议结构：

```go
type Service struct {
    store Store
    now   func() time.Time
}
```

`now` 使用函数而不是新增时钟框架，便于测试中固定时间。

Service 方法：

```go
Create(ctx context.Context, input CreateInput) (Credential, error)
Get(ctx context.Context, id int64) (Credential, error)
List(ctx context.Context, filter Filter) ([]Credential, error)
Update(ctx context.Context, id int64, input UpdateInput) (Credential, error)
Delete(ctx context.Context, id int64) error
```

创建时由 Service 设置 `CreatedAt` 和 `UpdatedAt`。更新时保留原 `CreatedAt`，只修改 `UpdatedAt`。

## 11. HTTP API

基础地址：

```text
http://127.0.0.1:8787
```

API 版本前缀：

```text
/api/v1
```

### 11.1 接口总览

| 方法 | 路径 | 功能 | 成功状态码 |
|---|---|---|---:|
| `GET` | `/healthz` | 健康检查 | `200` |
| `POST` | `/api/v1/credentials` | 新增账号 | `201` |
| `GET` | `/api/v1/credentials` | 获取账号列表 | `200` |
| `GET` | `/api/v1/credentials/{id}` | 获取单个账号 | `200` |
| `PUT` | `/api/v1/credentials/{id}` | 完整更新账号 | `200` |
| `DELETE` | `/api/v1/credentials/{id}` | 删除账号 | `204` |
| `GET` | `/api/v1/totp` | 获取已启用 2FA 账号的当前 TOTP 验证码 | `200` |

读取接口会返回密码、TOTP 密钥和恢复码字段，这是该工具用于读取凭据的核心行为。调用方必须把整个响应视为敏感信息。

### 11.2 健康检查

请求：

```http
GET /healthz
```

响应：

```json
{
  "status": "ok"
}
```

健康检查不返回版本、数据库路径或账号数量。

### 11.3 新增账号

请求：

```http
POST /api/v1/credentials
Content-Type: application/json
```

```json
{
  "provider": "github",
  "account": "user@example.com",
  "username": "octocat",
  "password": "plain-text-password"
}
```

响应状态码为 `201 Created`，响应体为完整记录：

```json
{
  "id": 1,
  "provider": "github",
  "account": "user@example.com",
  "username": "octocat",
  "password": "plain-text-password",
  "created_at": "2026-08-21T02:00:00Z",
  "updated_at": "2026-08-21T02:00:00Z"
}
```

### 11.4 获取账号列表

请求：

```http
GET /api/v1/credentials
```

可选过滤参数：

```text
GET /api/v1/credentials?provider=github
GET /api/v1/credentials?q=example.com
GET /api/v1/credentials?provider=github&q=octocat
```

`q` 对 `account` 和 `username` 执行包含查询。结果按 `id` 升序返回。

响应：

```json
{
  "items": [
    {
      "id": 1,
      "provider": "github",
      "account": "user@example.com",
      "username": "octocat",
      "password": "plain-text-password",
      "created_at": "2026-08-21T02:00:00Z",
      "updated_at": "2026-08-21T02:00:00Z"
    }
  ]
}
```

本地账号数量预计较少，第一版不实现分页。

### 11.5 获取单个账号

请求：

```http
GET /api/v1/credentials/1
```

存在时返回完整记录；不存在时返回 `404 Not Found`。

### 11.6 更新账号

`PUT` 使用完整替换语义，四个业务字段都应出现在请求中：

```http
PUT /api/v1/credentials/1
Content-Type: application/json
```

```json
{
  "provider": "github",
  "account": "new@example.com",
  "username": "new-name",
  "password": "new-password"
}
```

成功时返回更新后的完整记录；不存在时返回 `404 Not Found`。

### 11.7 删除账号

请求：

```http
DELETE /api/v1/credentials/1
```

成功时返回 `204 No Content`，响应体为空；不存在时返回 `404 Not Found`。

### 11.8 错误格式

所有 JSON 错误使用统一结构：

```json
{
  "error": {
    "code": "invalid_request",
    "message": "account is required"
  }
}
```

错误映射：

| 状态码 | `code` | 使用场景 |
|---:|---|---|
| `400` | `invalid_request` | JSON 错误、ID 错误、字段校验失败 |
| `404` | `not_found` | 账号记录不存在 |
| `405` | `method_not_allowed` | HTTP 方法不支持 |
| `415` | `unsupported_media_type` | 写接口不是 `application/json` |
| `500` | `internal_error` | 未预期的数据库或服务器错误 |

`500` 响应不得向调用方暴露 SQL、文件路径或堆栈信息。详细错误只能记录在本地日志中，并且日志仍不得包含密码。

### 11.9 JSON 请求处理要求

- 限制请求体大小，例如 64 KiB。
- 拒绝未知 JSON 字段，避免调用方拼错字段却误以为写入成功。
- 一个请求体只能包含一个 JSON 对象，后面不能附加第二个对象。
- 响应必须设置 `Content-Type: application/json; charset=utf-8`。

## 12. 项目目录

```text
OmniCred/
├── main.go                         # Wails 入口和前端资源嵌入
├── wails.json                      # Wails 构建与开发配置
├── build/
│   ├── appicon.svg                 # 应用图标源文件
│   ├── appicon.png                 # Wails 应用图标
│   ├── windows/                    # Windows manifest 和图标资源
│   └── bin/OmniCred.exe            # 构建产物，不提交 Git
├── internal/
│   ├── apiserver/
│   │   ├── server.go               # 本地 HTTP Server 生命周期
│   │   └── server_test.go
│   ├── desktop/
│   │   └── app.go                  # Wails 生命周期和单实例行为
│   ├── credential/
│   │   ├── model.go                # 领域模型和输入类型
│   │   ├── store.go                # Store 接口和公共错误
│   │   ├── service.go              # 业务逻辑
│   │   └── service_test.go
│   ├── httpapi/
│   │   ├── handler.go              # HTTP handler
│   │   ├── middleware.go           # 日志、恢复等少量中间件
│   │   ├── response.go             # JSON 和错误响应
│   │   └── handler_test.go
│   ├── sqlite/
│   │   ├── store.go                # SQLite Store 实现
│   │   ├── store_test.go
│   │   ├── migrate.go              # migration 执行器
│   │   └── migrations/
│   │       └── 001_create_credentials.sql
├── frontend/
│   ├── src/
│   │   ├── api/                     # API client、请求和响应类型
│   │   ├── components/
│   │   │   ├── credentials/         # 账号列表、卡片和表单
│   │   │   └── ui/                  # 按需添加的 shadcn/ui 组件
│   │   ├── hooks/                   # 页面级复用 hooks
│   │   ├── lib/                     # query client、格式化等小型工具
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.json
│   ├── dist/                        # 嵌入 EXE 的 Vite 构建产物
│   └── vite.config.ts
├── tools/
│   ├── linecheck/main.go            # 400 行限制检查工具
│   └── e2echeck/main.go             # 运行中 API 的端到端检查
├── data/
│   └── .gitkeep
├── .gitignore
├── go.mod
├── go.sum
├── README.md
└── ARCHITECTURE.md
```

不要增加 `controllers`、`repositories`、`usecases` 等内容重复的目录。第一版仅保留确实存在的职责；当文件接近 350 行时，在当前职责内部按功能拆分文件，而不是新增没有语义的架构层。

## 13. 程序启动与关闭

桌面程序固定使用 API 端口 `8787`，只提供可选数据库参数：

```text
--db D:\private\omnicred.db
```

没有指定 `--db` 时，数据库默认位于 `%APPDATA%\OmniCred\omnicred.db`。监听主机固定为 `127.0.0.1`，不提供修改地址或端口的参数。启动流程：

```text
双击 OmniCred.exe
   ↓
Wails 单实例检查
   ↓
创建并初始化数据库
   ↓
执行 migration
   ↓
创建 Store、Service、HTTP Handler
   ↓
创建独立桌面窗口
   ↓
OnStartup 监听 127.0.0.1:8787
   ↓
WebView 加载嵌入的 React 资源
   ↓
用户关闭窗口
   ↓
OnShutdown 停止 API 并关闭数据库
```

如果端口 `8787` 已被其他程序占用，桌面程序显示原生错误对话框并退出。再次双击正在运行的 OmniCred 时，不创建第二个进程，而是恢复并显示已有窗口。

## 14. 实现顺序

### 阶段一：项目骨架和数据模型

工作内容：

1. 初始化 Go module。
2. 创建目录结构。
3. 定义 `Credential`、输入类型、`Filter` 和 `Store`。
4. 增加 `.gitignore`，忽略 `data/*.db`、`data/*.db-*` 和构建产物。

验证：

```text
go build ./...
```

### 阶段二：SQLite Store

工作内容：

1. 接入纯 Go SQLite 驱动。
2. 实现数据库打开和 migration。
3. 实现 Store 的五个方法。
4. 使用临时目录中的数据库编写 Store 测试。

验证：

- 创建记录后能按 ID 读取。
- 更新后 `created_at` 不变，`updated_at` 改变。
- 删除后再次读取返回 `ErrNotFound`。
- provider 和关键词过滤正确。
- 关闭并重新打开数据库后数据仍然存在。

### 阶段三：Service

工作内容：

1. 实现输入校验和 provider 规范化。
2. 设置创建、更新时间。
3. 保持存储错误语义。

验证：

- 缺少必填字段时不调用 Store。
- 密码首尾空格被完整保留。
- provider 被保存为小写。
- 更新不存在的记录返回 `ErrNotFound`。

### 阶段四：HTTP API

工作内容：

1. 注册路由。
2. 实现严格 JSON 解析。
3. 实现统一成功和错误响应。
4. 增加不包含敏感数据的请求日志。

验证：

- 使用 `httptest` 覆盖所有接口。
- 覆盖错误 JSON、未知字段、错误 ID、记录不存在和错误 HTTP 方法。
- 检查测试日志中不出现测试密码。

### 阶段五：React 管理界面

工作内容：

1. 使用 React、TypeScript 和 Vite 初始化 `frontend`。
2. 配置 Tailwind CSS 和按需使用的 shadcn/ui 组件。
3. 实现账号列表、搜索、筛选、新增、编辑、复制和删除交互。
4. 接入 TanStack Query、React Hook Form 和 Zod。
5. 实现明暗主题、响应式布局和可访问交互。

验证：

- 375px、768px、1024px 和 1440px 宽度下没有水平溢出。
- 键盘能够完成搜索、新增、编辑、显示密码、复制和删除。
- 密码默认隐藏，浏览器存储和控制台中没有账号数据。
- 前端测试和生产构建通过。

### 阶段六：Wails 桌面集成、启动、关闭和文档

工作内容：

1. 将 Vite 生产构建输出到 `frontend/dist` 并嵌入 Wails。
2. 完成根目录 `main.go` 的依赖装配和桌面窗口配置。
3. 用 Wails `OnStartup`、`OnShutdown` 管理本地 API 生命周期。
4. 增加单实例锁、应用图标和用户配置目录数据库路径。
5. 编写 README 启动、桌面构建和 API 示例。
6. 增加 `tools/linecheck`、`tools/e2echeck` 并接入验收命令。
7. 完成桌面 EXE 和本地 API 端到端测试。

验证：

```text
npm --prefix frontend ci
npm --prefix frontend test
go run ./tools/linecheck
go test ./...
go vet ./...
wails build -clean -o OmniCred.exe
```

## 15. 必须覆盖的测试

| 测试层级 | 必须覆盖的内容 |
|---|---|
| Service 单元测试 | 必填字段、长度限制、provider 规范化、密码空格保留、时间更新 |
| SQLite 测试 | Create、Get、List、Update、Delete、NotFound、数据持久化 |
| HTTP 测试 | 六个接口、状态码、JSON 格式、错误映射、Content-Type |
| React 组件测试 | 列表、表单校验、显示密码、复制、Dialog 和错误状态 |
| 响应式与可访问性测试 | 键盘操作、焦点、对比度、375px 到桌面布局 |
| 安全回归测试 | 固定监听地址、Wails CORS 白名单、日志不包含密码、拒绝非 JSON 写请求 |
| 桌面集成测试 | 独立窗口、WebView2 加载、单实例唤醒、窗口关闭后 API 停止 |
| 端到端测试 | 启动桌面 EXE 并通过外部 API 完成一次完整 CRUD |
| 工程约束测试 | 所有受管代码文件总行数不超过 400 |

测试密码应使用明显的假数据，例如 `test-password-do-not-use`。

## 16. 手工验收示例

构建并启动桌面程序：

```powershell
wails build -clean -o OmniCred.exe
.\build\bin\OmniCred.exe --db .\data\manual-test.db
```

验收时必须确认：

- 出现标题为 `OmniCred` 的独立窗口。
- 没有打开默认浏览器。
- 第二次启动 EXE 后仍只有一个 OmniCred 主进程。
- 关闭窗口后 `127.0.0.1:8787` 不再监听。

创建 GitHub 账号：

```powershell
$body = @{
    provider = "github"
    account = "user@example.com"
    username = "octocat"
    password = "test-password-do-not-use"
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8787/api/v1/credentials" -ContentType "application/json" -Body $body
```

读取所有账号：

```powershell
curl.exe http://127.0.0.1:8787/api/v1/credentials
```

读取单个账号：

```powershell
curl.exe http://127.0.0.1:8787/api/v1/credentials/1
```

更新账号：

```powershell
curl.exe -X PUT http://127.0.0.1:8787/api/v1/credentials/1 -H "Content-Type: application/json" -d '{"provider":"github","account":"new@example.com","username":"new-name","password":"new-test-password"}'
```

删除账号：

```powershell
curl.exe -X DELETE http://127.0.0.1:8787/api/v1/credentials/1
```

## 17. 可扩展方式

扩展应沿已有边界进行：

| 未来需求 | 扩展方式 |
|---|---|
| 增加新的账号平台 | 直接写入新的 `provider`，无需修改代码 |
| 增加备注或网址字段 | 新增 migration，并同步修改模型和 API |
| 更换数据库 | 新增一个 `credential.Store` 实现 |
| 增加密码加密 | 在写入 Store 前加密、读取后解密，或提供加密 Store 实现 |
| 增加 API token | 在 HTTP 层增加认证中间件 |
| 增加 CLI | CLI 调用 Service 或本地 HTTP API，不复制业务规则 |
| 增加更多前端页面 | 继续调用现有 `/api/v1` API；确有多页面时再引入路由 |

不应把 GitHub、Google 等平台各自实现为插件或独立服务。只有当不同平台需要主动调用第三方 API 时，才考虑平台适配器。

## 18. 完成定义

第一版只有满足以下条件才算完成：

1. 能创建、读取、更新和删除账号记录。
2. 能保存并过滤 GitHub、Google、ChatGPT 和 OmniCred 记录。
3. 重启程序后记录仍然存在。
4. 服务默认且只能监听 `127.0.0.1`。
5. 数据库和日志中没有意外的额外密码副本。
6. 数据库文件不会被 Git 跟踪。
7. 所有自动化测试通过。
8. React 页面可以完成完整 CRUD、搜索、复制和密码显隐。
9. 双击 `OmniCred.exe` 打开独立窗口且不启动浏览器。
10. 第二次启动只唤醒已有窗口，关闭窗口后本地 API 停止。
11. 前端生产构建、`go vet ./...` 和 Wails 桌面构建通过。
12. `go run ./tools/linecheck` 通过，没有受管代码文件超过 400 行。
13. README 中包含桌面运行命令、数据风险说明、构建方式和 API 示例。

## 19. 架构决策摘要

- 选择单体进程，因为这是单机小型工具。
- 选择 SQLite，因为它比 JSON 文件更可靠，同时不需要独立服务。
- 选择标准库 HTTP，因为接口数量少，不需要框架。
- 选择 React 19、Vite、Tailwind CSS 和 shadcn/ui，因为需要现代、可维护且可定制的本地管理界面。
- 选择稳定版 Wails v2 承载 React，使用户双击一个 EXE 即可获得独立桌面窗口。
- Wails 使用系统 WebView2 渲染嵌入的前端资源，不打开默认浏览器。
- REST API 仍固定监听 `127.0.0.1:8787`，供桌面 WebView 和本地工具共同使用。
- 使用单实例锁避免两个进程同时占用 API 端口和数据库。
- 只抽象 Store，因为它是最明确、最可能变化的边界。
- provider 使用普通字符串，因为平台列表应由数据扩展，而不是由代码枚举扩展。
- 第一版返回明文密码，因为这是明确需求；同时严格限制监听地址和日志内容。
- 对所有手写代码执行 400 行硬限制，并在约 350 行时按职责拆分。
- 不实现未要求的认证、加密和同步，以保持第一版可以快速验证。
