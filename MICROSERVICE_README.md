# Golang 微服务示例：Nacos + Nats

这是一个使用 Golang 实现的微服务示例项目，集成了 Nacos（服务发现和配置管理）和 Nats（消息队列）。

## 项目结构

```
.
├── producer/          # 生产者服务 - 定期发布消息到 NATS
├── consumer/          # 消费者服务 - 订阅并处理来自 NATS 的消息
├── gateway/           # API 网关 - 提供 REST API 接口
├── config/            # 配置文件
├── docker-compose.yml # Docker Compose 配置
└── README.md          # 项目文档
```

## 功能特性

### 1. Nacos 服务发现
- 所有微服务在启动时自动注册到 Nacos
- 支持服务健康检查
- 动态服务发现和负载均衡

### 2. NATS 消息队列
- 生产者服务定期发布消息
- 消费者服务订阅并处理消息
- 支持发布-订阅模式

### 3. 微服务组件
- **Producer Service (端口 8081)**: 每 5 秒向 NATS 发布一条消息
- **Consumer Service (端口 8082)**: 订阅并打印接收到的消息
- **API Gateway (端口 8080)**: 提供 HTTP API 接口

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- Docker 和 Docker Compose（用于运行 Nacos 和 NATS）

### 1. 启动基础设施

首先启动 Nacos 和 NATS 服务：

```bash
docker-compose up -d
```

等待服务启动完成（约 30 秒），可以通过以下命令检查状态：

```bash
docker-compose ps
```

访问服务管理界面：
- Nacos 控制台: http://localhost:8848/nacos （用户名/密码: nacos/nacos）
- NATS 监控: http://localhost:8222

### 2. 启动微服务

#### 方式一：使用 go run（开发环境）

在不同的终端窗口中分别运行：

```bash
# 终端 1: 启动消费者服务
cd consumer
go mod download
go run main.go

# 终端 2: 启动生产者服务
cd producer
go mod download
go run main.go

# 终端 3: 启动 API 网关
cd gateway
go mod download
go run main.go
```

#### 方式二：编译后运行

```bash
# 编译所有服务
cd producer && go build -o producer && cd ..
cd consumer && go build -o consumer && cd ..
cd gateway && go build -o gateway && cd ..

# 运行服务
./consumer/consumer &
./producer/producer &
./gateway/gateway &
```

### 3. 测试服务

#### 查看已注册的服务

```bash
curl http://localhost:8080/api/services
```

#### 通过 API 发布消息

```bash
curl -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello from API Gateway!"}'
```

#### 健康检查

```bash
curl http://localhost:8080/health
```

## API 接口

### API Gateway (端口 8080)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/publish | 发布消息到 NATS |
| GET | /api/services | 查看所有注册的服务 |
| GET | /health | 健康检查 |

### 示例请求

**发布消息：**
```bash
curl -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "测试消息"}'
```

**查看服务列表：**
```bash
curl http://localhost:8080/api/services | jq
```

## 架构说明

### 服务注册与发现流程

1. 各服务启动时，通过 Nacos SDK 注册到 Nacos 服务器
2. 注册信息包括服务名称、IP、端口、元数据等
3. Gateway 可以通过 Nacos 查询其他服务的实例信息

### 消息流程

1. Producer 每 5 秒自动生成一条消息并发布到 NATS（主题：microservice.events）
2. Consumer 订阅该主题，接收并处理消息
3. Gateway 提供 HTTP 接口，允许外部通过 API 发布消息

### 服务交互图

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Producer  │      │   Gateway   │      │  Consumer   │
│   (8081)    │      │   (8080)    │      │   (8082)    │
└──────┬──────┘      └──────┬──────┘      └──────┬──────┘
       │                    │                    │
       │  Register          │  Register          │  Register
       ├───────────────────────────────────────────────────┐
       │                    │                    │          │
       │                    │                    │     ┌────▼─────┐
       │                    │                    │     │   Nacos  │
       │                    │                    │     │  (8848)  │
       │                    │                    │     └──────────┘
       │                    │                    │
       │  Publish           │  Publish           │  Subscribe
       ├────────────────────┴────────────────────┴─────────┐
       │                                                    │
       │                                               ┌────▼─────┐
       └───────────────────────────────────────────────►   NATS   │
                                                       │  (4222)  │
                                                       └──────────┘
```

## 配置说明

配置文件位于 `config/config.yml`：

- `nacos.server_addr`: Nacos 服务器地址
- `nacos.port`: Nacos 端口（默认 8848）
- `nats.url`: NATS 连接 URL
- `nats.subject`: 消息主题名称

## 日志说明

每个服务会输出详细的日志信息：

- Producer: 显示发布的每条消息
- Consumer: 显示接收到的每条消息
- Gateway: 显示 API 请求和处理结果

## 停止服务

### 停止微服务
使用 Ctrl+C 停止各个服务进程

### 停止基础设施
```bash
docker-compose down
```

## 技术栈

- **Golang**: 编程语言
- **Nacos**: 服务注册与发现、配置管理
- **NATS**: 消息队列（发布-订阅模式）
- **Docker**: 容器化部署

## 依赖项

主要 Go 依赖：
- `github.com/nacos-group/nacos-sdk-go`: Nacos Go SDK
- `github.com/nats-io/nats.go`: NATS Go 客户端

## 常见问题

### 1. 服务无法注册到 Nacos

确保 Nacos 服务已启动并可访问：
```bash
curl http://localhost:8848/nacos/v1/console/health/liveness
```

### 2. 消息没有被消费

检查 NATS 连接是否正常：
```bash
curl http://localhost:8222/varz
```

### 3. 端口被占用

修改各服务的端口号或停止占用端口的进程。

## 扩展建议

1. **配置中心**: 将配置从代码中分离到 Nacos 配置中心
2. **负载均衡**: 启动多个服务实例，实现负载均衡
3. **链路追踪**: 集成 OpenTelemetry 或 Jaeger
4. **监控告警**: 添加 Prometheus + Grafana
5. **API 网关增强**: 添加认证、限流、熔断等功能

## 许可证

MIT License

## 联系方式

- Email: 893400722@qq.com
- GitHub: @chengtb
