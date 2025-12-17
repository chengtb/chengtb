# 快速开始指南 (Quick Start Guide)

## 5 分钟快速运行微服务

### 步骤 1: 启动基础设施 (约 30 秒)

```bash
# 使用 Docker Compose 启动 Nacos 和 NATS
docker-compose up -d

# 等待服务就绪
sleep 15
```

### 步骤 2: 在三个终端窗口中启动微服务

**终端 1 - 启动消费者 (Consumer):**
```bash
cd consumer
go run main.go
```

**终端 2 - 启动生产者 (Producer):**
```bash
cd producer
go run main.go
```

**终端 3 - 启动网关 (Gateway):**
```bash
cd gateway
go run main.go
```

### 步骤 3: 验证服务运行

```bash
# 查看所有注册的服务
curl http://localhost:8080/api/services

# 通过 API 发送消息
curl -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello Microservices!"}'
```

### 预期结果

1. **消费者终端**将显示：
```
Received message: Message #1 from producer at 2025-12-17T06:21:00Z
Received message: Message #2 from producer at 2025-12-17T06:21:05Z
Received message: Hello Microservices!
```

2. **生产者终端**将显示：
```
Published: Message #1 from producer at 2025-12-17T06:21:00Z
Published: Message #2 from producer at 2025-12-17T06:21:05Z
```

3. **网关终端**将显示：
```
API Gateway is running on http://127.0.0.1:8080
Published message via API: Hello Microservices!
```

### 访问管理界面

- **Nacos 控制台**: http://localhost:8848/nacos
  - 用户名: `nacos`
  - 密码: `nacos`
  - 可以看到 3 个注册的服务

- **NATS 监控**: http://localhost:8222
  - 查看连接数和消息统计

## 使用 Makefile (推荐)

```bash
# 启动基础设施
make docker-up

# 在三个不同的终端中运行
make run-consumer  # 终端 1
make run-producer  # 终端 2  
make run-gateway   # 终端 3

# 测试 API
make test-api

# 停止基础设施
make docker-down
```

## 故障排查

### 问题: 端口被占用
```bash
# 检查端口占用
lsof -i :8080  # Gateway
lsof -i :8081  # Producer
lsof -i :8082  # Consumer
lsof -i :4222  # NATS
lsof -i :8848  # Nacos
```

### 问题: 服务无法连接到 Nacos 或 NATS
```bash
# 检查 Docker 容器状态
docker-compose ps

# 查看容器日志
docker-compose logs nacos
docker-compose logs nats

# 重启基础设施
docker-compose restart
```

## 架构图

```
┌─────────────────────────────────────────────────────────┐
│                   用户/客户端                            │
└─────────────────┬───────────────────────────────────────┘
                  │ HTTP REST API
                  ▼
         ┌────────────────┐
         │  API Gateway   │ :8080
         │  (gateway)     │
         └────┬───────┬───┘
              │       │
    ┌─────────┘       └────────┐
    │                          │
    │ NATS Publish      Service Discovery
    │                          │
    ▼                          ▼
┌─────────┐            ┌────────────┐
│  NATS   │ :4222      │   Nacos    │ :8848
│  消息队列 │            │  服务注册   │
└────┬────┘            └─────┬──────┘
     │                       │
     │ Pub/Sub        Register/Discover
     │                       │
┌────┴─────────────┬─────────┴─────┐
│                  │               │
▼                  ▼               ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Producer │  │ Consumer │  │ Gateway  │
│  :8081   │  │  :8082   │  │  :8080   │
└──────────┘  └──────────┘  └──────────┘
```

---

详细文档请参考: [MICROSERVICE_README.md](./MICROSERVICE_README.md)
