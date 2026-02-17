# UIUC 留学生接机排班后端

基于 Go 的接机排班系统后端，面向 UIUC 留学生接机业务流程。

## 当前功能范围

- 微信 `code` 登录并签发 JWT
- 通过微信 `getuserphonenumber` 绑定手机号
- 学生提交/修改接机需求
- 管理端司机与班次管理
- 管理端事务化分配学生到班次（含软超载 warning）
- 班次发布流程
- 航班同步定时任务占位（未配置时安全跳过）

## 技术栈

- Go + Gin
- GORM + MySQL 8.0
- Uber Fx
- JWT 鉴权
- robfig/cron v3

## 当前生效 API

基础路径：`/api/v1`

- `GET /health`
- `POST /auth/login`
- `POST /auth/bind-phone`（需 JWT）
- `POST /student/requests`（student）
- `GET /student/requests/my`（student）
- `PUT /student/requests/:id`（student）
- `GET /admin/drivers`（admin）
- `POST /admin/drivers`（admin）
- `GET /admin/shifts/dashboard`（admin）
- `GET /admin/requests/pending`（admin）
- `POST /admin/shifts`（admin）
- `POST /admin/shifts/:id/assign-student`（admin）
- `POST /admin/shifts/:id/remove-student`（admin）
- `POST /admin/shifts/:id/assign-staff`（admin）
- `POST /admin/shifts/:id/publish`（admin）

OpenAPI 文档源文件：[api/openapi.yaml](api/openapi.yaml)

## 快速开始

1）启动 MySQL（或使用已有 MySQL 8.0）

```bash
docker compose up -d
```

2）配置环境变量

```bash
cp env.example .env
```

然后把 `.env` 中变量导入当前 shell（或直接在系统环境变量中设置）。

3）可选：使用配置模板生成运行配置

- 复制 [files/config.template.yaml](files/config.template.yaml) 为 `files/config.yaml`
- 按需修改端口、地址、CORS、releaseMode

4）启动服务

```bash
go run app.go
```

## 配置说明

### 环境变量

完整示例见 [env.example](env.example)，关键变量：

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`, `JWT_EXPIRE_HOURS`, `JWT_ISSUER`
- `WECHAT_APPID`, `WECHAT_SECRET`
- `WECHAT_MCH_ID`, `WECHAT_MCH_KEY`, `WECHAT_NOTIFY_URL`
- `CRYPTO_KEY`
- `FLIGHT_API_URL`（可选，不配置时航班同步任务会跳过）

### 文件配置

模板文件：[files/config.template.yaml](files/config.template.yaml)

运行时文件（可选）：`files/config.yaml`

## 测试

```bash
go test ./...
```

覆盖率示例：

```bash
go test ./... -coverprofile=coverage_all -covermode=atomic
go tool cover -func coverage_all
```

## 说明

- 仓库中仍保留部分历史的 registration/order/payment 模块。
- 当前主流程已由 scheduler 路由接管，以上 API 为当前生效接口。