# 快速入门指南 (Quick Start Guide)

## 前提条件 (Prerequisites)

### 1. 安装 Go
确保已安装 Go 1.19 或更高版本:
```bash
go version
```

### 2. 启动 NATS Server

**方式一: 使用 Docker (推荐)**
```bash
docker run -d --name nats-server -p 4222:4222 nats:latest
```

**方式二: 本地安装**
- macOS: `brew install nats-server`
- Linux: 下载二进制文件 https://github.com/nats-io/nats-server/releases
- Windows: 下载安装包

启动 NATS:
```bash
nats-server
```

验证 NATS 运行:
```bash
# 查看 NATS 运行在端口 4222
netstat -an | grep 4222
```

## 运行示例 (Run Examples)

### 基础示例 (Basic Example)

展示核心功能: 事件发布、多回调订阅、异步执行

```bash
# 克隆仓库
git clone https://github.com/chengtb/chengtb.git
cd chengtb

# 安装依赖
go mod download

# 运行基础示例
go run examples/basic/main.go
```

**预期输出:**
```
✅ Connected to NATS server

📝 Registering callbacks for 'user.created' event...
📝 Registering callback for 'order.placed' event...

🚀 Publishing 'user.created' event...
📧 [Callback 1] Sending welcome email for event: evt-001
📊 [Callback 2] Updating analytics for event: evt-001
🔔 [Callback 3] Sending notification for event: evt-001
   ✅ Email sent successfully
   ✅ Analytics updated successfully
   ✅ Notification sent successfully

🚀 Publishing 'order.placed' event...
💳 [Order Callback] Processing payment for event: evt-002
   ✅ Payment processed successfully

✅ All events published and callbacks executed successfully!
🎉 Event-Driven Architecture demo completed!
```

### 高级示例 (Advanced Example)

展示错误处理、超时管理、工作流链

```bash
go run examples/advanced/main.go
```

**预期输出:**
```
✅ Connected to NATS server for advanced example

📝 Registering callbacks for 'payment.processed' event...

🚀 Publishing 'payment.processed' event...
✅ [Success Callback] Processing event: evt-payment-001
❌ [Error Callback] Simulating error for event: evt-payment-001
⏱️  [Timeout Callback] Starting long operation for event: evt-payment-001
   ✅ Long operation completed

🔗 Demonstrating callback chaining for order workflow...

🚀 Publishing 'order.workflow' event...
1️⃣  Validating order for event: evt-order-workflow-001
   ✅ Order validated
2️⃣  Processing payment for event: evt-order-workflow-001
   ✅ Payment processed
3️⃣  Updating inventory for event: evt-order-workflow-001
   ✅ Inventory updated
4️⃣  Sending confirmation for event: evt-order-workflow-001
   ✅ Confirmation sent

✅ Advanced example completed!
```

## 创建自己的应用 (Create Your Own Application)

### 第 1 步: 创建项目
```bash
mkdir my-eda-app
cd my-eda-app
go mod init my-eda-app
```

### 第 2 步: 添加依赖
```bash
go get github.com/chengtb/chengtb
go get github.com/nats-io/nats.go
```

### 第 3 步: 创建 main.go

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/chengtb/chengtb/pkg/eda/event"
    "github.com/chengtb/chengtb/pkg/eda/eventbus"
    "github.com/chengtb/chengtb/pkg/eda/publisher"
    "github.com/chengtb/chengtb/pkg/eda/subscriber"
)

func main() {
    // 1. 创建事件总线
    config := eventbus.Config{
        NATSUrl: "nats://localhost:4222",
    }
    
    bus, err := eventbus.NewEventBus(config)
    if err != nil {
        log.Fatalf("Failed to create event bus: %v", err)
    }
    defer bus.Close()
    
    fmt.Println("✅ Connected to NATS")

    // 2. 创建发布者和订阅者
    pub := publisher.NewPublisher(bus)
    sub := subscriber.NewSubscriber(bus)

    // 3. 定义回调函数
    myCallback := func(ctx context.Context, evt event.Event) error {
        fmt.Printf("📨 Received event: %s (type: %s)\n", 
            evt.GetID(), evt.GetType())
        fmt.Printf("   Payload: %+v\n", evt.GetPayload())
        return nil
    }

    // 4. 订阅事件
    if err := sub.SubscribeWithCallback("my.event", myCallback); err != nil {
        log.Fatalf("Failed to subscribe: %v", err)
    }

    // 等待订阅生效
    time.Sleep(100 * time.Millisecond)

    // 5. 发布事件
    evt := event.NewBaseEvent(
        "evt-001",
        "my.event",
        map[string]string{
            "message": "Hello EDA!",
        },
    )

    if err := pub.Publish(evt); err != nil {
        log.Fatalf("Failed to publish: %v", err)
    }

    // 等待回调执行
    time.Sleep(500 * time.Millisecond)
    fmt.Println("✅ Done!")
}
```

### 第 4 步: 运行应用
```bash
go run main.go
```

## 常见场景 (Common Scenarios)

### 场景 1: 用户注册流程

```go
// 定义回调
sendWelcomeEmail := func(ctx context.Context, evt event.Event) error {
    // 发送欢迎邮件
    return nil
}

createUserProfile := func(ctx context.Context, evt event.Event) error {
    // 创建用户配置
    return nil
}

notifyAdmins := func(ctx context.Context, evt event.Event) error {
    // 通知管理员
    return nil
}

// 注册回调
sub.SubscribeWithCallback("user.registered", sendWelcomeEmail)
sub.SubscribeWithCallback("user.registered", createUserProfile)
sub.SubscribeWithCallback("user.registered", notifyAdmins)

// 发布事件
userEvent := event.NewBaseEvent("evt-001", "user.registered", userData)
pub.Publish(userEvent)
```

### 场景 2: 订单处理流程

```go
// 工作流步骤
validateOrder := func(ctx context.Context, evt event.Event) error { /*...*/ }
reserveInventory := func(ctx context.Context, evt event.Event) error { /*...*/ }
processPayment := func(ctx context.Context, evt event.Event) error { /*...*/ }
sendConfirmation := func(ctx context.Context, evt event.Event) error { /*...*/ }

// 批量注册
callbacks := []callback.CallbackFunc{
    validateOrder,
    reserveInventory,
    processPayment,
    sendConfirmation,
}

sub.SubscribeWithMultipleCallbacks("order.created", callbacks)
```

### 场景 3: 实时通知系统

```go
// 通知回调
notifyCallback := func(ctx context.Context, evt event.Event) error {
    // 根据事件类型发送不同通知
    switch evt.GetType() {
    case "comment.added":
        // 发送评论通知
    case "like.received":
        // 发送点赞通知
    case "mention.received":
        // 发送提及通知
    }
    return nil
}

// 订阅多个事件
sub.SubscribeMultiple(
    []string{"comment.added", "like.received", "mention.received"},
    func(e event.Event) error {
        return notifyCallback(context.Background(), e)
    },
)
```

## 故障排查 (Troubleshooting)

### 问题 1: 无法连接到 NATS

**错误信息:**
```
Failed to create event bus: failed to connect to NATS: ...
```

**解决方案:**
1. 确认 NATS server 正在运行: `docker ps` 或 `ps aux | grep nats`
2. 检查端口 4222 是否被占用: `netstat -an | grep 4222`
3. 检查防火墙设置
4. 尝试使用 `127.0.0.1` 替代 `localhost`

### 问题 2: 回调没有执行

**可能原因:**
1. 订阅在发布之后 - 添加 `time.Sleep()` 等待订阅生效
2. 事件类型不匹配 - 检查事件类型字符串
3. NATS 连接断开 - 检查连接状态

**解决方案:**
```go
// 等待订阅生效
time.Sleep(100 * time.Millisecond)

// 检查连接
if !bus.IsConnected() {
    log.Fatal("NATS connection lost")
}
```

### 问题 3: 回调执行出错

**调试技巧:**
```go
callback := func(ctx context.Context, evt event.Event) error {
    // 添加 panic 恢复
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Callback panic: %v", r)
        }
    }()
    
    // 添加日志
    log.Printf("Processing event: %s", evt.GetID())
    
    // 业务逻辑
    if err := doSomething(); err != nil {
        log.Printf("Error: %v", err)
        return err
    }
    
    return nil
}
```

## 性能优化 (Performance Tips)

### 1. 复用连接
```go
// ✅ 好的做法: 复用 EventBus
var bus *eventbus.EventBus

func init() {
    bus, _ = eventbus.NewEventBus(config)
}

// ❌ 避免: 每次创建新连接
func publishEvent() {
    bus, _ := eventbus.NewEventBus(config)  // 不推荐
    // ...
}
```

### 2. 批量操作
```go
// ✅ 批量注册回调
callbacks := []callback.CallbackFunc{cb1, cb2, cb3}
sub.SubscribeWithMultipleCallbacks("event.type", callbacks)

// ❌ 避免: 逐个注册
sub.SubscribeWithCallback("event.type", cb1)
sub.SubscribeWithCallback("event.type", cb2)
sub.SubscribeWithCallback("event.type", cb3)
```

### 3. 控制并发
```go
// 使用 semaphore 控制并发数
sem := make(chan struct{}, 10)  // 最多 10 个并发

callback := func(ctx context.Context, evt event.Event) error {
    sem <- struct{}{}  // 获取信号
    defer func() { <-sem }()  // 释放信号
    
    // 处理事件
    return nil
}
```

## 下一步 (Next Steps)

1. 📖 阅读 [ARCHITECTURE.md](ARCHITECTURE.md) 了解架构细节
2. 💡 查看 [README.md](README.md) 获取完整文档
3. 🔧 根据需求定制和扩展
4. 🚀 部署到生产环境

## 获取帮助 (Get Help)

- 📧 Email: 893400722@qq.com
- 🐛 Issues: https://github.com/chengtb/chengtb/issues
- 💬 Discussions: https://github.com/chengtb/chengtb/discussions

## 许可证 (License)

MIT License - 可自由使用于商业和个人项目
