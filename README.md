# OmniCred

OmniCred 是一个双击即可运行的本地桌面账号密码管理器。桌面窗口使用 Wails + React 构建，后端使用 Go，数据保存在当前用户目录下的 SQLite 文件中。

> [!WARNING]
> 当前版本按照需求不加密密码。数据库中的密码是明文数据，任何能够读取数据库文件或访问本机 API 的程序都可能获取密码。请勿将数据库复制到不可信位置。

## 功能

- 独立 Windows 桌面窗口，不打开浏览器。
- 保存任意平台的登录账号、用户名和密码。
- 新增、搜索、筛选、编辑和删除账号。
- 新增、重命名和删除账号平台；重命名会同步更新关联账号。
- 根据邮箱在本地生成 GitHub 用户名和强密码，确认后再加入账号池。
- 粘贴带字段标签、`账号----密码` 或 `账号----密码----用户名` 格式的文本，支持单条填入和多行批量导入。
- 显示、隐藏和复制密码。
- React + shadcn/ui 风格的响应式明暗主题界面。
- 保留 REST API，供本地脚本和其他程序调用。
- SQLite 单文件持久化。
- 单实例运行；再次双击会唤醒已有窗口。
- 设置页可查看运行状态、迁移数据库位置并检查 GitHub Releases 更新。
- 安装版可从设置页确认后启动 Windows 卸载程序；便携版会禁用此入口。

## 直接运行

开发构建的桌面程序位于：

```text
build/bin/OmniCred.exe
```

双击 `OmniCred.exe` 即可打开独立窗口。也可以在 PowerShell 中运行：

```powershell
.\build\bin\OmniCred.exe
```

程序关闭时，SQLite 连接和本地 API 会一起停止。

## 数据位置

默认数据库位于当前 Windows 用户的配置目录：

```text
%APPDATA%\OmniCred\omnicred.db
```

如需指定数据库文件，可以使用：

```powershell
.\build\bin\OmniCred.exe --db D:\private\omnicred.db
```

数据库文件、账号、密码、2FA 密钥和恢复码都不会上传到云端。当前数据库未加密，请像保护密码一样保护 TOTP 密钥与恢复码。

也可以在应用左侧打开“设置”，选择新的 `.db` 文件位置。应用会在线复制完整数据库并保留旧文件，重启 OmniCred 后使用新位置。保存的路径配置位于：

```text
%APPDATA%\OmniCred\config.json
```

## 本地 API

桌面程序运行时会同时启动：

```text
http://127.0.0.1:8787
```

API 固定监听 `127.0.0.1`，不会监听局域网或公网地址。

| 方法 | 路径 | 功能 |
|---|---|---|
| `GET` | `/healthz` | 健康检查 |
| `POST` | `/api/v1/credentials` | 新增账号 |
| `GET` | `/api/v1/credentials` | 读取账号列表 |
| `GET` | `/api/v1/credentials/{id}` | 读取单个账号 |
| `PUT` | `/api/v1/credentials/{id}` | 更新账号 |
| `DELETE` | `/api/v1/credentials/{id}` | 删除账号 |
| `GET` | `/api/v1/totp` | 读取已启用 2FA 账号的当前 TOTP 验证码 |
| `POST` | `/api/v1/platforms` | 新增平台 |
| `GET` | `/api/v1/platforms` | 读取平台及关联账号数量 |
| `PUT` | `/api/v1/platforms/{id}` | 重命名平台及关联账号 |
| `DELETE` | `/api/v1/platforms/{id}` | 删除没有关联账号的平台 |
| `GET` | `/api/v1/settings/status` | 读取版本、数据路径和运行状态 |
| `PUT` | `/api/v1/settings/storage` | 复制数据库并设置下次启动位置 |
| `GET` | `/api/v1/settings/updates` | 检查 GitHub 最新稳定 Release |

创建账号：

```powershell
$body = @{
    provider = "github"
    account = "user@example.com"
    username = "octocat"
    password = "test-password-do-not-use"
    totp_secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
    recovery_codes = @("alpha-bravo", "charlie-delta")
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8787/api/v1/credentials" -ContentType "application/json" -Body $body
```

读取和筛选账号：

```powershell
curl.exe http://127.0.0.1:8787/api/v1/credentials
curl.exe "http://127.0.0.1:8787/api/v1/credentials?provider=github"
curl.exe "http://127.0.0.1:8787/api/v1/credentials?q=octocat"
```

API 的读取响应包含明文密码，调用方必须把整个响应视为敏感数据。

## 开发环境

- Go 1.25 或更高版本。
- Node.js 22 或更高版本。
- npm 10 或更高版本。
- Wails v2.14。
- Windows WebView2 Runtime。

安装与项目匹配的 Wails CLI：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@v2.14.0
```

检查桌面构建环境：

```powershell
wails doctor
```

## 开发模式

```powershell
wails dev
```

该命令会启动 Vite、Go 后端和可热更新的 OmniCred 桌面窗口。开发模式不会自动打开浏览器，并使用独立的 `%APPDATA%\OmniCred-Dev\` 配置与数据库目录以及 `127.0.0.1:8788` API 端口，不会读写生产版数据。

## 构建桌面 EXE

```powershell
wails build -clean -o OmniCred.exe
```

构建结果：

```text
build/bin/OmniCred.exe
```

React、字体、应用图标和 Go 后端都会打包进 EXE。运行时依赖 Windows 自带或已安装的 WebView2 Runtime，不需要单独启动 Node.js 或浏览器。

## 构建中英文安装包

先安装 NSIS，然后运行：

```powershell
wails build -clean -nsis -o OmniCred.exe
```

`build/bin` 中会同时生成独立 EXE 和 Windows 安装器。安装器启动时可选择 English 或简体中文，卸载界面会沿用安装语言。

## GitHub Releases

项目发布仓库：

```text
https://github.com/mibgb65-cloud/OmniCred
```

Windows 安装器会分别让用户选择程序安装目录和数据存储目录；首次启动时，数据目录选择会写入现有应用配置，重装不会覆盖已有设置。

每次发布前，先创建与标签同名的说明文件，例如 [`docs/releases/v0.1.0.md`](./docs/releases/v0.1.0.md)。推送符合 `vMAJOR.MINOR.PATCH` 的标签（例如 `v0.2.0`）后，[`.github/workflows/release.yml`](./.github/workflows/release.yml) 会读取对应说明作为 Release Notes，并在 Windows runner 上运行测试、构建 EXE 和双语安装器、生成 SHA-256 校验文件以及创建 GitHub Release。缺少对应说明文件时发布会失败。不要把本地数据库或真实密码提交到仓库。

## 验证

前端测试与构建：

```powershell
cd frontend
npm test
npm run build
cd ..
```

Go 测试和静态检查：

```powershell
go test ./...
go vet ./...
```

检查代码文件行数：

```powershell
go run ./tools/linecheck
```

当 OmniCred 正在运行时，可以执行 API 端到端测试：

```powershell
go run ./tools/e2echeck
```

项目维护的 `.go`、`.ts`、`.tsx`、`.js`、`.jsx`、`.css` 和 `.sql` 文件不得超过 400 行。注释和空行也计入总行数；依赖目录、锁文件和构建产物除外。

## 项目结构

```text
main.go                 Wails 桌面程序入口和前端资源嵌入
internal/apiserver/     本地 REST API 生命周期
internal/desktop/       桌面窗口启动、关闭和单实例行为
internal/credential/    领域模型、校验和 Store 接口
internal/httpapi/       REST API 和 HTTP 中间件
internal/sqlite/        SQLite Store 与 migration
frontend/               React、TypeScript、Vite 和 UI 组件
build/                  应用图标、Windows 构建资源和最终 EXE
tools/linecheck/        400 行限制检查工具
tools/e2echeck/         本地 API 端到端检查工具
```

详细设计参见 [ARCHITECTURE.md](./ARCHITECTURE.md)。
