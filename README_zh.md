# 接机管理系统

一个基于 Go + Gin + GORM 的微信小程序接机管理系统后端服务。

## 功能特性

- 🔐 微信小程序一键登录
- 📝 接机报名管理
- 💰 微信支付集成
- 📋 订单状态跟踪
- 📢 消息板公告
- 👥 司机分配管理
- 🔒 JWT 认证授权
- 📊 完整的 API 文档

## 技术栈

- **语言**: Go 1.25+
- **框架**: Gin (HTTP Router)
- **数据库**: MySQL 8.0+
- **ORM**: GORM
- **依赖注入**: Uber FX
- **日志**: Zap
- **配置**: Viper
- **认证**: JWT

## 项目结构

```
pickup/
├── api/                    # API 文档
│   └── openapi.yaml       # OpenAPI 3.0 规范
├── files/                  # 配置文件
│   ├── config.yaml        # 主配置文件
│   └── logs/              # 日志文件
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   ├── service/          # 业务逻辑层
│   └── utils/            # 工具函数
├── migrations/           # 数据库迁移
├── pkg/                  # 公共包
│   ├── config/          # 配置包
│   ├── server/          # 服务器包
│   └── zap/             # 日志包
├── tests/                # 测试文件
├── go.mod               # Go 模块文件
├── go.sum               # Go 依赖校验
└── README.md            # 项目说明
```

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- 微信小程序 AppID 和 AppSecret

### 安装依赖

```bash
go mod tidy
```

### 配置环境变量

复制环境变量示例文件：

```bash
cp env.example .env
```

编辑 `.env` 文件，填入实际配置：

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=pickup

# 微信小程序配置
WECHAT_APPID=your_wechat_appid
WECHAT_SECRET=your_wechat_secret
WECHAT_MCH_ID=your_merchant_id
WECHAT_MCH_KEY=your_merchant_key
WECHAT_NOTIFY_URL=https://yourdomain.com/api/v1/pay/notify

# JWT配置
JWT_SECRET=your_jwt_secret_key_should_be_long_and_random
JWT_EXPIRE_HOURS=24
JWT_ISSUER=pickup

# 加密配置
CRYPTO_KEY=your_crypto_key_32_characters_long
```

### 数据库初始化

1. 创建数据库：

```sql
CREATE DATABASE pickup CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 运行数据库迁移：

```bash
# 执行迁移脚本
mysql -u root -p pickup < migrations/001_initial_schema.up.sql
```

### 启动服务

```bash
go run app.go
```

服务将在 `http://localhost:8080` 启动。

## API 文档

启动服务后，可以通过以下方式查看 API 文档：

- OpenAPI 文档：`http://localhost:8080/api/openapi.yaml`
- 健康检查：`http://localhost:8080/api/v1/health`

### 主要 API 端点

#### 认证相关

- `POST /api/v1/auth/wechat/login` - 微信登录
- `GET /api/v1/auth/me` - 获取当前用户信息

#### 报名管理

- `POST /api/v1/registrations` - 创建报名
- `GET /api/v1/registrations/:id` - 获取报名详情
- `PUT /api/v1/registrations/:id` - 更新报名
- `DELETE /api/v1/registrations/:id` - 删除报名
- `GET /api/v1/registrations/my` - 获取我的报名列表

#### 订单管理

- `POST /api/v1/orders` - 创建订单
- `GET /api/v1/orders/:id` - 获取订单详情
- `GET /api/v1/orders/my` - 获取我的订单列表

#### 支付管理

- `POST /api/v1/pay/prepare` - 准备支付
- `POST /api/v1/pay/notify` - 支付回调

#### 管理端

- `POST /api/v1/admin/orders/:id/notify` - 通知订单

## 测试

### 运行全部测试并查看覆盖率

```powershell
powershell -File run_tests.ps1
```

该脚本会运行全部测试，按函数显示彩色覆盖率报告，并输出总覆盖率。

### 运行全部测试（命令行）

```bash
go test ./... -v
```

### 生成覆盖率报告

```bash
go test ./... -coverprofile=coverage/coverage -count=1
go tool cover -func coverage/coverage
```

### 测试覆盖率概览

| 包名                 | 覆盖率   |
| -------------------- | -------- |
| internal/config      | 39.3%    |
| internal/handler     | 98.4%    |
| internal/middleware   | 86.2%    |
| internal/model       | 100.0%   |
| internal/repository  | 92.1%    |
| internal/service     | 85.3%    |
| internal/utils       | 89.4%    |
| pkg/config           | 83.3%    |
| pkg/server           | 63.8%    |
| pkg/zap              | 100.0%   |
| **总计**             | **89.2%**|

处理器测试和 Mock 位于 `tests/` 目录（auth、order、payment、registration、notice）。Postman 集合文件：[tests/postman_collection.json](tests/postman_collection.json)。

## 部署

### Docker 部署

1. 构建镜像：

```bash
docker build -t pickup-api .
```

2. 运行容器：

```bash
docker run -d \
  --name pickup-api \
  -p 8080:8080 \
  --env-file .env \
  pickup-api
```

### 生产环境配置

1. 设置 `server.releaseMode: true`
2. 配置正确的数据库连接
3. 设置微信支付相关配置
4. 配置 HTTPS 证书
5. 设置日志轮转和监控

## 开发指南

### 添加新的 API

1. 在 `internal/model/` 中定义数据模型
2. 在 `internal/repository/` 中实现数据访问层
3. 在 `internal/service/` 中实现业务逻辑
4. 在 `internal/handler/` 中实现 HTTP 处理器
5. 在 `internal/handler/router.go` 中注册路由
6. 更新 `api/openapi.yaml` 文档
7. 编写测试用例

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 编写完整的注释和文档
- 保持函数简洁，单一职责
- 使用有意义的变量和函数名

## 常见问题

### Q: 如何获取微信小程序的 code？

A: 在小程序端调用 `wx.login()` 获取 code，然后调用 `wx.getPhoneNumber()` 获取手机号授权 code。

### Q: 支付回调如何处理？

A: 微信支付成功后会自动调用 `/api/v1/pay/notify` 接口，系统会自动更新订单状态。

### Q: 如何添加新的用户角色？

A: 在 `internal/model/user.go` 中的 `UserRole` 类型中添加新角色，并更新数据库迁移脚本。

### Q: 如何配置微信支付？

A: 需要先在微信商户平台配置支付参数，然后在环境变量中设置相应的配置。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](LICENSE) 文件。

## 联系方式

- 项目维护者：Pickup Team
- 项目地址：https://github.com/ZocP/wechat

## 更新日志

### v1.0.0 (2025-10-11)

- ✨ 初始版本发布
- 🔐 实现微信登录功能
- 📝 实现报名管理功能
- 💰 集成微信支付
- 📋 实现订单管理
- 📢 实现消息板功能
- 🔒 实现 JWT 认证
- 📊 完整的 API 文档
- 🧪 单元测试和集成测试

