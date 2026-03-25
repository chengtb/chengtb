# DDD Go 微服务基础框架

基于 **Go** 语言，采用 **领域驱动设计（DDD）** 规范搭建的微服务基础框架。

- 🌐 **API 网关**：使用 [Gin](https://github.com/gin-gonic/gin) 提供 RESTful HTTP 接口
- 📨 **服务通信**：使用 [NATS](https://nats.io/) Request-Reply 模式实现微服务间通信，消息载荷采用 **[Protocol Buffers](https://protobuf.dev/)** 二进制编码
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
├── proto/
│   └── user.proto  # Protobuf Schema 定义（请求/响应消息 + Reply 信封）
├── internal/
│   ├── domain/
│   │   └── user/   # 领域层：实体、值对象、仓储接口、领域服务、领域事件
│   ├── application/
│   │   └── user/   # 应用层：命令/查询处理（CQRS）、应用服务
│   ├── infrastructure/
│   │   ├── messaging/    # NATS 连接、发布者（Client）、订阅者（Subscriber）
│   │   ├── persistence/  # 仓储实现（内存）
│   │   └── transport/    # Gin 引擎初始化及中间件
│   ├── interfaces/
│   │   ├── api/    # HTTP 路由和处理器
│   │   └── dto/    # 请求/响应数据传输对象
│   └── pb/
│       └── user.pb.go    # 由 protoc 自动生成的 Go 代码（勿手动修改）
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

## Protobuf 消息格式

NATS 通道上的所有消息均使用 **Protocol Buffers（proto3）** 进行二进制序列化，Schema 定义位于 [`proto/user.proto`](proto/user.proto)，对应的 Go 代码由 `protoc` 自动生成到 [`internal/pb/user.pb.go`](internal/pb/user.pb.go)。

### 消息定义

#### 请求消息（客户端 → 微服务）

| NATS Subject    | 消息类型               | 字段                        |
|----------------|----------------------|----------------------------|
| `user.create`  | `CreateUserRequest`  | `id`, `name`, `email`      |
| `user.get`     | `GetUserRequest`     | `id`                       |
| `user.list`    | `ListUsersRequest`   | （无字段）                   |
| `user.update`  | `UpdateUserRequest`  | `id`, `name`, `email`      |
| `user.delete`  | `DeleteUserRequest`  | `id`                       |

#### 响应消息（微服务 → 客户端）

| 消息类型     | 字段                                                                   |
|------------|----------------------------------------------------------------------|
| `User`     | `id`, `name`, `email`, `created_at_unix`（纳秒时间戳）, `updated_at_unix` |
| `UserList` | `users`（`User` 列表）                                                  |
| `Empty`    | （无字段，用于无返回值操作，如删除）                                         |

### Reply 信封协议

每一条 NATS 响应都包裹在统一的 `Reply` 信封中：

```protobuf
message Reply {
  int32  error_code    = 1;  // 0 = 成功，非 0 = 错误
  string error_message = 2;  // 错误时携带可读描述
  bytes  data          = 3;  // 成功时携带 proto 序列化的响应消息
}
```

**成功响应**示例（以 `GetUser` 为例）：

```
Reply {
  error_code    = 0
  error_message = ""
  data          = <proto.Marshal(User{id:"1", name:"Alice", ...})>
}
```

**错误响应**示例：

```
Reply {
  error_code    = 1
  error_message = "user not found"
  data          = <空>
}
```

### 代码分层

```
proto/user.proto              ← Schema 权威来源（手动维护）
    │
    │ protoc 生成
    ▼
internal/pb/user.pb.go        ← 自动生成（勿手动修改）
    │
    ├── internal/infrastructure/messaging/subscriber.go
    │       proto.Unmarshal(msg.Data, &req)   ← 解码请求
    │       proto.Marshal(Reply{...})         ← 编码响应
    │
    └── internal/infrastructure/messaging/publisher.go
            proto.Marshal(req)                ← 编码请求
            proto.Unmarshal(reply.Data, &out) ← 解码响应
```

### 如何重新生成 Protobuf 代码

修改 `proto/user.proto` 之后，执行以下命令重新生成 Go 代码：

**1. 安装工具链（仅需一次）**

```bash
# 安装 protoc 编译器
# macOS
brew install protobuf
# Ubuntu / Debian
sudo apt-get install -y protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

**2. 生成代码**

```bash
# 在项目根目录执行
protoc \
  --go_out=. \
  --go_opt=module=github.com/chengtb/chengtb \
  proto/user.proto
```

生成的文件位于 `internal/pb/user.pb.go`，**请将其提交到版本库**，以避免使用者需要在本地安装 `protoc`。

### 扩展消息定义

如需新增消息（例如支持新的业务操作），步骤如下：

1. 在 `proto/user.proto` 中添加新的 `message` 定义
2. 运行上述 `protoc` 命令重新生成 `internal/pb/user.pb.go`
3. 在 `subscriber.go` 中注册新的 NATS Subject 和处理函数
4. 在 `publisher.go` 中添加对应的客户端方法

---

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
