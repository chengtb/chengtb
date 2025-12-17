# 故障排查指南 (Troubleshooting Guide)

## 常见问题和解决方案

### 1. Docker 相关问题

#### 问题: Docker Compose 无法启动
```bash
# 检查 Docker 是否运行
docker info

# 如果没有权限，添加当前用户到 docker 组
sudo usermod -aG docker $USER
newgrp docker

# 重启 Docker 服务
sudo systemctl restart docker
```

#### 问题: 容器端口冲突
```bash
# 检查端口占用
sudo lsof -i :8848  # Nacos
sudo lsof -i :4222  # NATS
sudo lsof -i :8080  # Gateway

# 停止占用端口的进程
sudo kill -9 <PID>

# 或修改 docker-compose.yml 中的端口映射
```

#### 问题: 容器启动失败
```bash
# 查看容器日志
docker-compose logs nacos
docker-compose logs nats

# 完全重建容器
docker-compose down -v
docker-compose up -d --force-recreate
```

### 2. Nacos 相关问题

#### 问题: 服务无法注册到 Nacos
**症状**: 日志显示 "failed to register service"

**解决方案**:
```bash
# 1. 确认 Nacos 已启动
curl http://localhost:8848/nacos/v1/console/health/liveness

# 2. 检查 Nacos 日志
docker-compose logs nacos | tail -50

# 3. 访问 Nacos 控制台确认
# http://localhost:8848/nacos (nacos/nacos)

# 4. 等待 Nacos 完全启动 (可能需要 30 秒)
sleep 30

# 5. 检查服务代码中的配置
# 确保 IP 和端口正确: 127.0.0.1:8848
```

#### 问题: Nacos 控制台无法访问
```bash
# 检查容器状态
docker-compose ps

# 如果容器正在重启，等待其稳定
watch docker-compose ps

# 查看详细日志
docker-compose logs -f nacos
```

#### 问题: 服务列表中看不到服务
1. 确保服务已启动并运行
2. 检查服务日志是否有注册成功的消息
3. 在 Nacos 控制台中，选择正确的命名空间 (public)
4. 刷新服务列表页面

### 3. NATS 相关问题

#### 问题: 无法连接到 NATS
**症状**: "failed to connect to NATS"

**解决方案**:
```bash
# 1. 检查 NATS 是否运行
curl http://localhost:8222/healthz

# 2. 检查 NATS 连接信息
curl http://localhost:8222/varz | jq '.'

# 3. 测试 NATS 连接
# 使用 nats-cli (需要安装)
nats server ls
```

#### 问题: 消息发布失败
```bash
# 1. 检查 NATS 监控页面
curl http://localhost:8222/connz | jq '.'

# 2. 查看当前订阅
curl http://localhost:8222/subsz | jq '.'

# 3. 重启 NATS 容器
docker-compose restart nats
```

#### 问题: 消费者收不到消息
1. **检查消费者是否正常订阅**:
   - 查看消费者日志: "Subscribed to subject: microservice.events"
   
2. **检查主题名称是否一致**:
   - Producer: `NatsSubject = "microservice.events"`
   - Consumer: `NatsSubject = "microservice.events"`
   
3. **确认消费者在生产者之前启动**

### 4. Go 编译问题

#### 问题: 缺少依赖包
```bash
# 在每个服务目录下执行
go mod download
go mod tidy
```

#### 问题: 版本冲突
```bash
# 清理缓存
go clean -modcache

# 重新下载依赖
go mod download

# 如果还有问题，删除 go.sum 重新生成
rm go.sum
go mod tidy
```

#### 问题: 编译错误
```bash
# 检查 Go 版本
go version  # 需要 1.21 或更高

# 更新 Go 版本
# 访问 https://golang.org/dl/

# 清理并重新编译
go clean
go build -v
```

### 5. 微服务运行问题

#### 问题: 端口已被占用
```bash
# 查看端口占用
lsof -i :8080  # Gateway
lsof -i :8081  # Producer
lsof -i :8082  # Consumer

# 结束占用端口的进程
kill -9 <PID>

# 或在代码中修改端口
# producer/main.go: ServicePort = 8081
# consumer/main.go: ServicePort = 8082
# gateway/main.go:  ServicePort = 8080
```

#### 问题: 服务启动后立即退出
```bash
# 检查依赖服务是否运行
docker-compose ps

# 查看完整的错误日志
go run main.go 2>&1 | tee service.log
```

#### 问题: 服务之间无法通信
1. **检查服务注册状态**:
   ```bash
   curl http://localhost:8080/api/services | jq '.'
   ```

2. **检查网络连接**:
   ```bash
   ping localhost
   telnet localhost 8848  # Nacos
   telnet localhost 4222  # NATS
   ```

### 6. API 测试问题

#### 问题: curl 命令失败
```bash
# 检查 Gateway 是否运行
curl http://localhost:8080/health

# 使用详细模式查看错误
curl -v http://localhost:8080/api/services

# 检查请求格式
curl -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "test"}' \
  -v
```

#### 问题: jq 命令未找到
```bash
# 安装 jq (JSON 处理工具)
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq

# 或者不使用 jq
curl http://localhost:8080/api/services
```

### 7. 日志和调试

#### 查看所有服务状态
```bash
# 基础设施
docker-compose ps
docker-compose logs --tail=50

# 微服务 (在各自的终端中)
# 观察日志输出，查找错误信息
```

#### 启用详细日志
在代码中修改日志级别 (临时调试):
```go
// 在 main.go 中添加
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

#### 常见错误消息

1. **"connection refused"**
   - 目标服务未启动
   - 端口号错误
   - 防火墙阻止连接

2. **"context deadline exceeded"**
   - 服务响应超时
   - 网络问题
   - 服务过载

3. **"failed to register service"**
   - Nacos 未就绪
   - 配置错误
   - 网络问题

### 8. 性能问题

#### 问题: 服务响应慢
```bash
# 检查系统资源
top
htop
docker stats

# 检查 NATS 性能
curl http://localhost:8222/varz | jq '.cpu, .mem'

# 检查 Nacos 性能
docker-compose logs nacos | grep -i "slow"
```

### 9. 清理和重置

#### 完全清理环境
```bash
# 停止所有服务
pkill -f "go run"  # 停止所有 Go 进程

# 清理 Docker
docker-compose down -v
docker system prune -f

# 清理临时文件
rm -rf /tmp/nacos

# 清理 Go 缓存
go clean -cache -modcache -testcache

# 重新开始
docker-compose up -d
sleep 15
# 然后重新启动微服务
```

## 调试技巧

### 1. 逐步验证
```bash
# 步骤 1: 验证基础设施
docker-compose ps
curl http://localhost:8848/nacos/
curl http://localhost:8222/healthz

# 步骤 2: 启动一个服务验证
cd consumer && go run main.go

# 步骤 3: 在 Nacos 控制台检查服务是否注册

# 步骤 4: 依次启动其他服务
```

### 2. 使用日志追踪问题
在每个服务的关键位置添加日志:
```go
log.Printf("DEBUG: Connecting to NATS at %s", natsURL)
log.Printf("DEBUG: Registering service %s at %s:%d", serviceName, ip, port)
```

### 3. 使用网络工具
```bash
# 查看所有网络连接
netstat -tulpn | grep -E "8080|8081|8082|8848|4222"

# 抓包分析
sudo tcpdump -i lo -n port 4222  # 监控 NATS 流量
```

## 获取帮助

如果以上方法都无法解决问题:

1. 查看项目 Issues: https://github.com/chengtb/chengtb/issues
2. 收集以下信息:
   - 操作系统和版本
   - Go 版本 (`go version`)
   - Docker 版本 (`docker --version`)
   - 完整的错误日志
   - 重现步骤

3. 联系维护者: 893400722@qq.com

## 附录: 有用的命令

```bash
# 查看所有 Go 进程
ps aux | grep "go run"

# 一键重启所有服务
make clean && make docker-down && make docker-up
sleep 15
# 然后在不同终端启动服务

# 快速测试
./test-api.sh

# 监控日志
tail -f /tmp/nacos/log/*.log
```
