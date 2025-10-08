# DNSMesh · 多提供商 DNS 管理平台

> Self-hosted domain governance for Cloudflare & Tencent Cloud DNS.

DNSMesh 是一套面向开发者的自托管域名解析运营平台，提供多云 DNS 资产的统一管理、智能分析与安全审计能力。目前原生支持 Cloudflare 与腾讯云 DNSPod，并通过可扩展的服务接口为后续接入更多云厂商预留空间。

## ✨ 功能亮点

- **多云接入**：内置 Cloudflare 与腾讯云 DNSPod 连接器，可在创建 Provider 时自动校验凭证并首轮同步所有解析记录。
- **凭证加密存储**：通过 AES-256-GCM 对 API Token、Secret 等敏感信息进行加密，密钥由 `ENCRYPTION_KEY` 环境变量提供。
- **服务器优先视图**：后端的智能分析服务会根据域名模式、CNAME 引用与 IP 复用情况自动归类服务器，前端以“服务器卡片 + 快速添加”方式呈现。
- **批量导入与重新分析**：连接器同步的记录可在导入向导中挑选批量入库；支持一键“重新分析”来重新同步所有 Provider 并刷新服务器分组。
- **精细化记录管理**：支持新建、编辑、隐藏（脱管）与删除解析记录；隐藏操作会保留数据库记录但停止纳管，便于回溯。
- **安全的会话机制**：使用基于 Cookie 的会话与中间件保护所有 `/api` 下敏感接口，并提供用户名、密码修改入口。
- **审计日志**：对 Provider 与解析记录的增删改、同步等操作留痕，可按资源类型、动作过滤并分页查询。
- **API 优先设计**：前端通过轻量的 Mithril 服务层消费 REST API，便于后续接入 CLI 或其他自动化工具。

## 🧱 系统架构

- **后端**：Go 1.21、Gin、GORM、PostgreSQL。入口位于 `backend/cmd/main.go`，核心逻辑分层于 `internal/{handlers,services,models,middleware,database}`，并使用 Cookie Session 存储登录态。
- **服务层**：`internal/services` 封装各云厂商 SDK；`AnalyzeDNSRecords` 用于模式识别与服务器分组，所有 Provider 共用 `DNSProvider` 接口实现统一的同步、增删改删除协议。
- **加密模块**：`pkg/crypto` 负责初始化与执行 AES-256-GCM 加解密，保证凭据落库前被加密。
- **前端**：基于 Vite 与 Mithril 构建的单页应用，组件划分为 `views`（页面）、`components`（弹窗/卡片）、`services/api.js`（API 适配层）以及 `styles/main.css`（全局样式）。
- **静态资源分发**：前端构建产物位于 `backend/public`，由 Gin 直接作为静态文件提供，实现“一体化”部署。

```
┌────────────┐      ┌────────────────┐
│  Frontend  │──XHR│  REST Gateway   │──┐
│  (Mithril) │     │ (Gin Handlers)  │  │
└────────────┘      └────────────────┘  │
        ▲                     │          │
        │ Cookies             ▼          ▼
 ┌────────────┐       ┌──────────────┐┌─────────────┐
 │User Session│◀────▶│ GORM + Postgres││Provider SDK│
 └────────────┘       └──────────────┘└─────────────┘
```

## 🚀 快速上手

### 环境依赖

- Go 1.21 或更高版本
- Node.js 18 及以上（推荐搭配 npm 9+）
- PostgreSQL 15（可通过 Docker Compose 提供）
- Docker / Docker Compose（可选，用于一体化启动与部署）

### 本地开发步骤

1. **克隆仓库**
   ```bash
   git clone <your-repo-url>
   cd dnsmesh
   ```

2. **启动数据库（可选）**
   ```bash
   docker-compose up -d postgres
   ```

3. **配置后端环境变量**
   ```bash
   cd backend
   cp .env.example .env
   # 按需修改数据库连接、SESSION_SECRET、ADMIN_*、ENCRYPTION_KEY 等变量
   ```

4. **运行后端 API**
   ```bash
   go mod tidy
   go run cmd/main.go
   ```
   服务默认监听 `http://localhost:8080`，首次启动会自动迁移数据库并创建默认管理员账号。

5. **运行前端开发服务器**
   ```bash
   cd ../frontend
   npm install
   npm run dev
   ```
   前端默认运行在 `http://localhost:3000`，通过 Vite 代理访问后端 API。

> 提示：项目根目录提供 `./start.sh` 作为便捷脚本，会顺序启动 Postgres、后端与前端（需本机已安装 Docker、Go、Node）。

### 构建生产版本

1. 在 `frontend` 目录执行 `npm run build`，构建产物会写入 `frontend/dist`。
2. 将 `frontend/dist` 内容拷贝至 `backend/public`（Vite 构建时也可以直接指定输出路径到 `../backend/public`）。
3. 准备生产环境 `.env` 并确保 `GIN_MODE=release`、`SESSION_SECRET` 与 `ENCRYPTION_KEY` 均为强随机值。
4. 运行后端二进制或使用下方 Docker Compose 模式。

## 🔧 环境变量

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | 后端监听端口 |
| `GIN_MODE` | `release` | Gin 运行模式（开发环境可设为 `debug`） |
| `DB_HOST` | `localhost` | PostgreSQL 主机地址 |
| `DB_PORT` | `5432` | PostgreSQL 端口 |
| `DB_USER` | `dnsmesh` | 数据库用户名 |
| `DB_PASSWORD` | _(空)_ | 数据库密码 |
| `DB_NAME` | `dnsmesh` | 数据库名称 |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL 模式 |
| `SESSION_SECRET` | `default-secret-change-in-production` | Cookie 会话加密密钥，请在生产环境务必替换 |
| `ADMIN_USERNAME` | `admin` | 首次启动创建的管理员用户名 |
| `ADMIN_PASSWORD` | `admin` | 首次启动创建的管理员密码（仅当数据库无用户时生效） |
| `ENCRYPTION_KEY` | _(必填)_ | 32 字节字符串，用于 AES-256-GCM 加密 Provider 凭据，未设置会导致应用启动失败 |

## 🐳 Docker Compose 部署

项目根目录提供 `docker-compose.yml`：

1. 修改 `docker-compose.yml` 中 `backend` 服务的环境变量，确保 `SESSION_SECRET`、`ADMIN_PASSWORD`、`ENCRYPTION_KEY` 等符合生产要求。
2. 构建前端并将产物放入 `backend/public`。
3. 执行：
   ```bash
   docker-compose up -d
   ```
   将启动 `postgres` 与 `backend` 两个服务，默认暴露端口 `5432` 与 `8080`。

## 🧪 测试与质量保障

- **Go 单元测试**：`cd backend && go test ./...`
- **烟雾测试脚本**：项目根目录的 `./test.sh` 会模拟登录并调用核心 API，便于快速验证部署可用性。
- **格式化**：提交前建议执行 `cd backend && go fmt ./...` 与前端的 `npm run lint`（若后续添加）。

## 🗂️ 目录结构速览

```
dnsmesh/
├── backend/
│   ├── cmd/                # 应用入口 (main.go)
│   ├── internal/
│   │   ├── handlers/       # HTTP 路由处理器
│   │   ├── services/       # 业务逻辑 & Provider 接入
│   │   ├── models/         # GORM 模型
│   │   ├── middleware/     # Gin 中间件
│   │   └── database/       # 连接、迁移、默认用户
│   ├── pkg/crypto/         # AES 加密工具
│   └── public/             # 前端打包产物
├── frontend/
│   ├── src/
│   │   ├── views/          # 页面视图 (Login, Dashboard)
│   │   ├── components/     # 弹窗、表单、向导
│   │   ├── services/       # API 请求封装
│   │   └── styles/         # Tailwind 风格的手写样式
│   ├── public/             # Vite 静态资源
│   └── vite.config.js
├── docker-compose.yml
├── start.sh                # 本地一键启动脚本
└── test.sh                 # 烟雾测试脚本
```

## 🌐 API 概览

所有除 `/api/auth/login` 以外的接口均需在成功登录后携带会话 Cookie (`withCredentials: true`) 才能访问。

### 认证
- `POST /api/auth/login`：用户名密码登录。
- `POST /api/auth/logout`：退出登录并清除会话。
- `GET /api/auth/user`：获取当前登录用户信息。
- `POST /api/auth/change-password`：修改密码。
- `POST /api/auth/change-username`：修改用户名并刷新会话。

### DNS 提供商
- `GET /api/providers`：获取 Provider 列表（敏感字段会被清空）。
- `POST /api/providers`：创建 Provider，会在落库前试连并加密凭据。
- `PUT /api/providers/:id`：更新 Provider 凭据并重新验证连通性。
- `DELETE /api/providers/:id`：删除 Provider 及其关联解析记录。
- `POST /api/providers/:id/sync`：同步指定 Provider 的全部解析记录并返回分析结果。

### DNS 记录
- `GET /api/records`：返回“服务器优先 + 未分组”结构的解析记录。
- `POST /api/records`：为已知服务器创建新的解析记录（自动推断 Zone 和 Provider）。
- `PUT /api/records/:id`：更新解析记录，若关键字段变化会同步至 Provider。
- `POST /api/records/:id/hide`：将记录标记为不再纳管（仅软删除）。
- `DELETE /api/records/:id`：从 Provider 与数据库双向删除（仅针对非服务器记录）。
- `POST /api/records/import`：批量导入同步结果中的记录。
- `POST /api/records/reanalyze`：重新同步所有 Provider 并刷新服务器建议。

### 审计日志
- `GET /api/audit-logs`：分页查询审计日志，支持 `resource_type`、`action`、`limit`、`offset` 参数过滤。

## 💡 前端交互要点

- **Provider Wizard**：两步式弹窗，先连接 Provider，再勾选同步记录；可预填建议的服务器名称与地域。
- **服务器卡片视图**：展示每台服务器的主记录、关联域名与快速操作（添加、隐藏、删除）。
- **未分组记录区**：按 Provider 聚合未能匹配服务器的记录，便于后续补充元数据或隐藏。
- **重新分析入口**：位于工具栏，可触发后端对所有 Provider 再次同步并更新服务器识别结果。
- **账户设置弹窗**：支持修改用户名与密码，更新后本地存储与会话信息会同步刷新。

## 🔐 安全实践

- 生产环境必须设置强随机的 `SESSION_SECRET` 与 32 字节 `ENCRYPTION_KEY`。
- 默认管理员仅在用户表为空时创建，部署后请立即修改用户名与密码或禁用默认账号。
- 建议在 HTTPS 环境下运行并将 `cookie` 的 `Secure` 选项设为 `true`（可在 `main.go` 中根据部署情况调整）。
- 对 Provider 凭据的备份需谨慎处理，避免将 `.env` 与数据库明文导出。

## 🗺️ 路线图

- [ ] 支持更多 DNS Provider（如阿里云、华为云等）。
- [ ] 细粒度的角色权限控制与多用户协作。
- [ ] 更丰富的报表与指标（历史变更、区域覆盖率）。
- [ ] Webhook / CLI 集成，用于自动化批量操作。

欢迎通过 Issue 或 Pull Request 提交建议与功能需求。

## 🤝 贡献指南

1. Fork 本仓库并创建特性分支：`git checkout -b feature/my-feature`。
2. 完成开发后运行必要的格式化与测试。
3. 提交 Pull Request 时请描述问题背景、解决方案以及测试情况。
4. 后端代码请保持与现有分层一致，前端组件命名遵循 PascalCase，函数使用 camelCase。

## 📄 许可证

项目尚未指定开源许可证。若计划公开发布，请在根目录添加如 MIT、Apache-2.0 或其他合适的许可证文件，并在此处更新说明。

## 🙏 致谢

- Cloudflare Go SDK 与腾讯云 DNSPod SDK 提供的 API 能力。
- Mithril 社区提供的轻量化前端框架。
- 以及所有愿意贡献改进 DNSMesh 的开发者。

---

如在部署或使用过程中遇到问题，欢迎提 Issue 讨论，也期待你的 Star 与贡献！
- [x] 前端项目初始化

### Phase 2: DNS Provider 集成 (进行中)
- [ ] Cloudflare API 集成
- [ ] 腾讯云 DNSPod API 集成
- [ ] Provider CRUD 功能
- [ ] 智能分析算法

### Phase 3: 核心功能
- [ ] DNS 记录管理
- [ ] 服务器分组展示
- [ ] 快速添加功能
- [ ] 批量导入记录

### Phase 4: 增强功能
- [ ] 操作历史记录
- [ ] 搜索和过滤
- [ ] 从 Provider 同步
- [ ] 错误处理优化

### Phase 5: 部署和优化
- [ ] 生产环境配置
- [ ] 性能优化
- [ ] 文档完善

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！
