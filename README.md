# DDD Go 微服务基础框架

基于 **Go** 语言，采用 **领域驱动设计（DDD）** 规范搭建的微服务基础框架。

- 🌐 **API 网关**：使用 [Gin](https://github.com/gin-gonic/gin) 提供 RESTful HTTP 接口
- 📨 **服务通信**：使用 [NATS](https://nats.io/) Request-Reply 模式实现微服务间通信
- 📝 **日志**：使用 [Zap](https://github.com/uber-go/zap) 结构化日志
- 🏗️ **架构**：严格遵循 DDD 分层架构（Domain → Application → Infrastructure → Interfaces）

---

## 项目结构

```
.
├── cmd/
│   ├── api/        # API 网关入口（Gin HTTP 服务器）
│   └── service/    # User 微服务入口（NATS 订阅者）
├── config/         # 环境变量配置
├── internal/
│   ├── domain/
│   │   └── user/   # 领域层：实体、值对象、仓储接口、领域服务、领域事件
│   ├── application/
│   │   └── user/   # 应用层：命令/查询处理（CQRS）、应用服务
│   ├── infrastructure/
│   │   ├── messaging/    # NATS 连接、发布者（Client）、订阅者（Subscriber）
│   │   ├── persistence/  # 仓储实现（内存）
│   │   └── transport/    # Gin 引擎初始化及中间件
│   └── interfaces/
│       ├── api/    # HTTP 路由和处理器
│       └── dto/    # 请求/响应数据传输对象
└── pkg/
    ├── errors/     # 统一错误类型与 HTTP 状态码映射
    └── logger/     # Zap 日志封装
```

## 架构说明

```
HTTP 请求
    │
    ▼
[Gin HTTP Server] (cmd/api)
    │  interfaces/api — 路由/处理器
    │  使用 NATS Client 发送 Request
    ▼
[NATS Broker]  ←→  Request-Reply 模式
    ▼
[User Microservice] (cmd/service)
    │  infrastructure/messaging — Subscriber
    │  application/user         — 应用服务（CQRS）
    │  domain/user              — 聚合根、领域服务
    │  infrastructure/persistence — 仓储实现
    ▼
[Repository]
```

## 快速开始

### 前置条件

- Go 1.25+
- [NATS Server](https://docs.nats.io/running-a-nats-service/introduction/installation)

### 1. 启动 NATS Server

```bash
nats-server
```

### 2. 启动 User 微服务

```bash
go run ./cmd/service
```

### 3. 启动 API 网关

```bash
go run ./cmd/api
```

### 4. 测试接口

```bash
# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"id":"1","name":"Alice","email":"alice@example.com"}'

# 查询所有用户
curl http://localhost:8080/api/v1/users

# 查询单个用户
curl http://localhost:8080/api/v1/users/1

# 更新用户
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated"}'

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/1

# 健康检查
curl http://localhost:8080/health
```

## 环境变量

| 变量名         | 默认值                    | 说明              |
|---------------|--------------------------|------------------|
| `HTTP_ADDR`   | `:8080`                  | API 监听地址      |
| `NATS_ADDR`   | `nats://localhost:4222`  | NATS 服务地址     |
| `LOG_LEVEL`   | `info`                   | 日志级别          |
| `SERVICE_NAME`| `user-service`           | 服务名称          |

## 运行测试

```bash
go test ./...
```

---

- 📫 联系方式：893400722@qq.com
