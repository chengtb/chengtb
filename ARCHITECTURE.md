# 架构设计文档 (Architecture Design)

## 系统架构概览

### 1. 整体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                          客户端/用户                               │
│                    (Web Browser, CLI, API Client)                │
└─────────────────────────────┬────────────────────────────────────┘
                              │
                              │ HTTP REST API
                              │
                  ┌───────────▼──────────┐
                  │                      │
                  │   API Gateway        │
                  │   Port: 8080         │
                  │                      │
                  └───────┬──────────┬───┘
                          │          │
            ┌─────────────┘          └─────────────┐
            │                                      │
            │ Publish Messages             Query Services
            │                                      │
        ┌───▼────────┐                    ┌───────▼──────┐
        │            │                    │              │
        │   NATS     │                    │    Nacos     │
        │  Message   │                    │   Service    │
        │   Queue    │                    │  Registry    │
        │  :4222     │                    │   :8848      │
        │            │                    │              │
        └───┬────────┘                    └───────┬──────┘
            │                                     │
            │ Pub/Sub                             │ Register/
            │                              Discover/Watch
            │                                     │
  ┌─────────┴──────────┬──────────────────────────┴─────────┐
  │                    │                                     │
┌─▼──────────┐   ┌────▼────────┐                  ┌─────────▼────┐
│            │   │             │                  │              │
│  Producer  │   │  Consumer   │                  │   Gateway    │
│  Service   │   │  Service    │                  │   Service    │
│  :8081     │   │  :8082      │                  │   :8080      │
│            │   │             │                  │              │
└────────────┘   └─────────────┘                  └──────────────┘
```

## 2. 组件详细说明

### 2.1 Producer Service (生产者服务)

**职责**:
- 定期生成业务消息
- 发布消息到 NATS 消息队列
- 向 Nacos 注册服务

**端口**: 8081

**关键功能**:
```go
// 服务注册
RegisterService(ip, port) -> Nacos

// 消息发布 (每 5 秒)
PublishMessage(message) -> NATS
```

**数据流**:
1. 启动时注册到 Nacos
2. 每 5 秒生成一条消息
3. 将消息发布到 NATS 主题 "microservice.events"
4. 发送心跳给 Nacos 保持注册状态

### 2.2 Consumer Service (消费者服务)

**职责**:
- 订阅 NATS 消息队列
- 处理接收到的消息
- 向 Nacos 注册服务

**端口**: 8082

**关键功能**:
```go
// 服务注册
RegisterService(ip, port) -> Nacos

// 订阅消息
Subscribe(subject) -> NATS
OnMessage(handler)
```

**数据流**:
1. 启动时注册到 Nacos
2. 订阅 NATS 主题 "microservice.events"
3. 接收并打印所有消息
4. 保持长连接监听新消息

### 2.3 API Gateway (API 网关)

**职责**:
- 提供 HTTP REST API 接口
- 转发客户端请求到 NATS
- 查询 Nacos 服务列表
- 向 Nacos 注册服务

**端口**: 8080

**API 端点**:

| 方法 | 路径 | 功能 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | /api/publish | 发布消息 | `{"content": "..."}` | `{"status": "success"}` |
| GET | /api/services | 查询服务 | - | `{"services": [...]}` |
| GET | /health | 健康检查 | - | `{"status": "healthy"}` |

**数据流**:
1. 启动时注册到 Nacos
2. 监听 HTTP 请求
3. 接收 POST /api/publish 请求，发布到 NATS
4. 接收 GET /api/services 请求，从 Nacos 查询
5. 返回 JSON 响应给客户端

### 2.4 Nacos (服务注册中心)

**职责**:
- 服务注册与发现
- 健康检查
- 配置管理（本示例未使用）

**端口**: 8848, 9848

**功能**:
- 服务注册: 接收服务实例信息
- 服务发现: 提供服务查询接口
- 健康检查: 监控服务实例状态
- Web 控制台: 可视化管理界面

**数据存储**:
- 内存模式（standalone）
- 可配置为数据库持久化

### 2.5 NATS (消息队列)

**职责**:
- 消息发布订阅
- 消息路由
- 高性能消息传递

**端口**: 4222 (client), 8222 (monitoring)

**特性**:
- 轻量级、高性能
- 发布-订阅模式
- 主题路由
- 支持 JetStream（持久化）

## 3. 交互流程

### 3.1 服务启动流程

```
1. 启动 Docker Compose
   └─> 启动 Nacos (等待就绪)
   └─> 启动 NATS

2. 启动 Consumer Service
   └─> 连接 NATS (:4222)
   └─> 订阅主题 "microservice.events"
   └─> 连接 Nacos (:8848)
   └─> 注册服务 "consumer-service"

3. 启动 Producer Service
   └─> 连接 NATS (:4222)
   └─> 连接 Nacos (:8848)
   └─> 注册服务 "producer-service"
   └─> 开始发布消息 (每 5 秒)

4. 启动 Gateway Service
   └─> 连接 NATS (:4222)
   └─> 连接 Nacos (:8848)
   └─> 注册服务 "api-gateway"
   └─> 启动 HTTP 服务器 (:8080)
```

### 3.2 消息发布流程

#### 场景 A: Producer 自动发布

```
Producer (定时器触发)
  │
  ├─> 生成消息: "Message #N from producer at <timestamp>"
  │
  ├─> 调用 NATS Publish()
  │       │
  │       └─> NATS Server 路由消息
  │                 │
  │                 └─> 推送给所有订阅者
  │                           │
  │                           └─> Consumer 收到消息
  │                                    │
  │                                    └─> 打印到日志
  │
  └─> 记录日志: "Published: ..."
```

#### 场景 B: 通过 API 发布

```
Client (curl/浏览器)
  │
  ├─> POST /api/publish
  │     Content-Type: application/json
  │     Body: {"content": "Hello"}
  │
  └─> Gateway 接收请求
        │
        ├─> 解析 JSON
        │
        ├─> 调用 NATS Publish()
        │       │
        │       └─> NATS Server 路由消息
        │                 │
        │                 └─> Consumer 收到消息
        │
        ├─> 返回响应: {"status": "success"}
        │
        └─> 记录日志
```

### 3.3 服务发现流程

```
Client
  │
  ├─> GET /api/services
  │
  └─> Gateway 处理请求
        │
        ├─> 查询 Nacos: producer-service
        │       │
        │       └─> 返回实例列表 [{ip, port, ...}]
        │
        ├─> 查询 Nacos: consumer-service
        │       │
        │       └─> 返回实例列表
        │
        ├─> 查询 Nacos: api-gateway
        │       │
        │       └─> 返回实例列表
        │
        ├─> 聚合结果
        │
        └─> 返回 JSON: {"services": [...], "count": 3}
```

## 4. 数据模型

### 4.1 服务注册数据

```json
{
  "serviceName": "producer-service",
  "ip": "127.0.0.1",
  "port": 8081,
  "weight": 10,
  "enable": true,
  "healthy": true,
  "ephemeral": true,
  "metadata": {
    "version": "1.0"
  }
}
```

### 4.2 消息格式

```
Producer 生成:
"Message #1 from producer at 2025-12-17T06:21:00Z"

API 发布:
客户端发送: {"content": "Hello World"}
NATS 传输: "Hello World"
```

### 4.3 API 响应格式

**发布消息响应**:
```json
{
  "status": "success",
  "message": "Message published successfully"
}
```

**服务列表响应**:
```json
{
  "services": [
    {
      "name": "producer-service",
      "instances": [
        {
          "instanceId": "...",
          "ip": "127.0.0.1",
          "port": 8081,
          "healthy": true,
          "metadata": {"version": "1.0"}
        }
      ]
    }
  ],
  "count": 3
}
```

## 5. 技术决策

### 5.1 为什么选择 Nacos?

- ✅ 开源、社区活跃
- ✅ 支持服务注册与发现
- ✅ 支持配置管理
- ✅ 提供 Web UI
- ✅ 与 Spring Cloud 生态集成好
- ✅ 国内文档丰富

**替代方案**: Consul, Eureka, Etcd

### 5.2 为什么选择 NATS?

- ✅ 轻量级、高性能
- ✅ 简单易用
- ✅ 支持多种消息模式
- ✅ 低延迟
- ✅ Go 原生支持

**替代方案**: RabbitMQ, Kafka, Redis Pub/Sub

### 5.3 为什么使用 Go?

- ✅ 高并发支持
- ✅ 编译快、部署简单
- ✅ 标准库强大
- ✅ 适合微服务开发
- ✅ 性能优秀

## 6. 扩展性设计

### 6.1 水平扩展

可以启动多个服务实例:

```bash
# Terminal 1: Consumer Instance 1
cd consumer && PORT=8082 go run main.go

# Terminal 2: Consumer Instance 2  
cd consumer && PORT=8083 go run main.go

# 负载会在实例间分配（需配置）
```

### 6.2 添加新服务

创建新服务的步骤:
1. 复制现有服务代码
2. 修改服务名称和端口
3. 实现业务逻辑
4. 注册到 Nacos
5. 通过 NATS 通信

### 6.3 配置管理

未来可以将配置移到 Nacos Config:
```go
// 从 Nacos 读取配置
config, err := configClient.GetConfig(vo.ConfigParam{
    DataId: "application.yml",
    Group:  "DEFAULT_GROUP",
})
```

## 7. 监控和运维

### 7.1 健康检查

- Nacos: `http://localhost:8848/nacos/v1/console/health/liveness`
- NATS: `http://localhost:8222/healthz`
- Gateway: `http://localhost:8080/health`

### 7.2 监控指标

**NATS 监控**:
```bash
curl http://localhost:8222/varz  # 服务器信息
curl http://localhost:8222/connz # 连接信息
curl http://localhost:8222/subsz # 订阅信息
```

**Nacos 监控**:
- Web UI: http://localhost:8848/nacos
- 查看服务列表、实例健康状态

### 7.3 日志

每个服务输出结构化日志:
```
2025/12/17 06:21:00 Service producer-service registered successfully at 127.0.0.1:8081
2025/12/17 06:21:05 Published: Message #1 from producer at 2025-12-17T06:21:05Z
```

## 8. 安全考虑

### 当前实现

- 本地开发环境
- 无认证/授权
- 明文通信

### 生产环境建议

1. **API Gateway**:
   - 添加 JWT 认证
   - 实现 API 限流
   - HTTPS/TLS

2. **Nacos**:
   - 启用认证
   - 配置命名空间隔离
   - 使用 MySQL 持久化

3. **NATS**:
   - 启用认证
   - TLS 加密
   - ACL 权限控制

## 9. 最佳实践

1. **服务注册**: 使用临时节点 (ephemeral: true)
2. **错误处理**: 实现重试机制和熔断器
3. **日志**: 统一日志格式，便于聚合分析
4. **配置**: 环境变量 > 配置文件 > 硬编码
5. **优雅关闭**: 正确处理 SIGTERM/SIGINT 信号

## 10. 参考资源

- [Nacos 官方文档](https://nacos.io/zh-cn/docs/what-is-nacos.html)
- [NATS 官方文档](https://docs.nats.io/)
- [Go 微服务最佳实践](https://github.com/golang-standards/project-layout)
