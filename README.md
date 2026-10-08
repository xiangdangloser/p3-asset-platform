# P3 · asset-platform（资产扫描与监控平台）

> 🛑 **本项目仅供网络安全学习、授权渗透测试及自有资产测绘使用。使用前请务必阅读文末的【法律声明与免责条款】。**

基于 Go 语言开发的资产扫描、指纹识别与异步任务调度平台。
本项目是「Go × 网安 12 个月双线计划」的核心大项目（P3），旨在将 P0（并发探测）和 P1（指纹识别）的 CLI 工具，升级为具备完整鉴权与数据持久化能力的企业级 Web 服务。

## ✨ 核心功能

- **用户认证与鉴权**：基于 JWT（JSON Web Token）的无状态身份认证，结合 bcrypt 密码哈希，保护敏感 API 不被未授权访问。
- **异步任务下发**：接收扫描请求后毫秒级返回 `task_id`，通过 Goroutine 在后台并发执行扫描，不阻塞主线程。
- **高并发扫描引擎**：内置 Worker Pool（生产者-消费者模型），支持超时重试与 User-Agent 伪装。
- **指纹识别引擎**：内置 20+ 条国内常见 OA、中间件与框架指纹规则（RuoYi、JeecgBoot、Nacos、Tengine、Nginx 等）。
- **数据持久化与冲突处理**：使用 GORM + SQLite（纯 Go 驱动 `glebarez/sqlite`），自动迁移表结构，并利用 `clause.OnConflict` 实现相同目标扫描结果的覆盖更新（免去频繁删库的烦恼）。
- **任务状态轮询**：提供 `/api/v1/tasks/:id` 接口，支持前端实时查询后台扫描进度。

## 🛠️ 技术栈

- **语言**：Go 1.26+
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：SQLite（纯 Go 驱动，无 CGO 依赖，跨平台无忧）
- **认证**：JWT (golang-jwt/jwt/v5)
- **密码安全**：bcrypt
- **架构**：标准 RESTful 分层架构（Controller / Service / Model / Middleware）

## 📁 项目结构

```text
p3-asset-platform/
├── controllers/       # 控制器：处理 HTTP 请求、参数绑定与响应
├── database/          # 数据库：GORM 初始化、自动迁移与重置
├── middlewares/       # 中间件：JWT 鉴权拦截器
├── models/            # 数据模型：Asset（资产）, Task（任务）, User（用户）
├── routes/            # 路由：API 分组与注册
├── services/          # 核心服务：并发扫描引擎与指纹识别规则库
├── main.go            # 入口：依赖组装与服务启动
├── go.mod             # 依赖管理
└── .gitignore         # Git 忽略规则（已忽略 *.db 与 *.exe）


## 🚀 快速开始

### 1. 编译与运行
```bash
git clone https://github.com/xiangdangloser/p3-asset-platform.git
cd p3-asset-platform
go mod tidy
go run main.go

服务默认监听 http://127.0.0.1:8080。首次启动会自动创建 assets.db 数据库并完成建表。

2. 数据库重置（开发环境专用）
如果你在开发过程中修改了 models 里的结构体字段（如修改了类型或删除了字段），GORM 的自动迁移可能无法平滑更新表结构。此时只需：

停止服务（Ctrl+C）。

手动删除项目根目录下的 assets.db 文件。

重新运行 go run main.go，程序会自动重建全新的表结构。

📡 API 接口说明
方法	路径	说明	鉴权
POST	/api/v1/register	注册新用户	❌
POST	/api/v1/login	登录并获取 JWT Token	❌
POST	/api/v1/scan	下发异步扫描任务	✅ Bearer Token
GET	/api/v1/tasks/:id	查询指定任务的状态与进度	✅ Bearer Token
GET	/api/v1/assets	获取已扫描的资产列表	✅ Bearer Token
快速测试流程（Apifox / Postman）：
调用 /register 注册账号（Body 传 {"username":"admin", "password":"123456"}）。

调用 /login 获取 token。

在 Apifox 的 Auth 标签页选择 Bearer Token，填入刚才的 token。

调用 /scan 传参：{"targets": ["https://www.baidu.com", "https://www.taobao.com"]}。

立即获得 task_id（毫秒级响应），然后调用 /tasks/1 查看进度。

扫描完成后，调用 /assets 查看持久化在数据库中的资产与指纹结果。

⚠️ 已知局限性（Known Limitations）
当前指纹引擎基于简单字符串匹配（strings.Contains），尚未引入正则表达式，可能在某些定制页面上产生误报（False Positive）。

当前异步任务的状态仅支持内存级轮询，若服务重启，running 状态的任务会中断（缺少断点续扫机制）。

SQLite 数据库适合单机低并发场景，若需集群化部署，需要更换为 PostgreSQL 或 MySQL。

🗺️ 未来规划（Roadmap）
□ P2 深度联动：将本地引擎规则转换为 Nuclei 模板并提交 PR，反向增加识别规则。
□ 容器化部署：提供 Dockerfile 与 docker-compose.yml，实现一键在 Linux 服务器启动。
□ 前端可视化：开发基于 Vue3 + Element Plus 的简单管理界面，展示扫描进度与资产列表。
□ 主动指纹识别：引入 favicon.ico 哈希匹配、主动路径探测等增强识别准确率。
□ CI/CD：配置 GitHub Actions 实现自动化编译与测试。
⚠️ 法律声明与免责条款
合法授权：本工具仅供网络安全学习、授权渗透测试及企业自有资产测绘使用。

严禁未授权扫描：严禁任何个人或组织使用本工具对未授权的第三方目标进行扫描、探测、爆破或攻击。

法律红线：批量扫描未授权目标，在国内属于《中华人民共和国网络安全法》和《中华人民共和国刑法》第 285 条的射程范围。

免责声明：使用者在使用本工具时需自行承担因操作不当或非法使用而产生的所有法律责任。作者不对任何滥用行为及由此产生的后果负责。

当你克隆、下载或运行本项目时，即代表你已阅读并同意上述条款。

## 📄 License
本项目采用 MIT License 开源协议。详情请参阅 [LICENSE](LICENSE) 文件。
