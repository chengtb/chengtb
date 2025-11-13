# Event-Driven Architecture with Golang + NATS

一个基于 Golang 和 NATS 的事件驱动架构(EDA)实现，具有强大的回调功能。

## 🌟 特性

- ✅ **基于 NATS 的事件总线**: 使用 NATS 作为消息中间件，提供高性能的事件传输
- ✅ **灵活的事件系统**: 支持自定义事件类型和负载
- ✅ **回调机制**: 支持为事件注册多个回调函数
- ✅ **异步执行**: 回调函数异步并发执行，提高性能
- ✅ **发布/订阅模式**: 标准的发布订阅模式实现
- ✅ **错误处理**: 完善的错误处理和回调执行管理
- ✅ **易于使用**: 简洁的 API 设计，易于集成

## 📋 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [架构概述](#架构概述)
- [核心组件](#核心组件)
- [使用示例](#使用示例)
- [API 文档](#api-文档)

## 🚀 安装

### 前置要求

- Go 1.19 或更高版本
- NATS Server (用于消息传输)

### 安装 NATS Server

使用 Docker 运行 NATS:

```bash
docker run -p 4222:4222 nats:latest
```

或者下载安装: https://docs.nats.io/running-a-nats-service/introduction/installation

### 安装项目依赖

```bash
go get github.com/chengtb/chengtb
```

## ⚡ 快速开始

### 基础示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chengtb/chengtb/pkg/eda/callback"
    "github.com/chengtb/chengtb/pkg/eda/event"
    "github.com/chengtb/chengtb/pkg/eda/eventbus"
    "github.com/chengtb/chengtb/pkg/eda/publisher"
    "github.com/chengtb/chengtb/pkg/eda/subscriber"
)

func main() {
    // 创建事件总线
    config := eventbus.Config{
        NATSUrl: "nats://localhost:4222",
    }
    bus, err := eventbus.NewEventBus(config)
    if err != nil {
        log.Fatal(err)
    }
    defer bus.Close()

    // 创建发布者和订阅者
    pub := publisher.NewPublisher(bus)
    sub := subscriber.NewSubscriber(bus)

    // 定义回调函数
    myCallback := func(ctx context.Context, evt event.Event) error {
        fmt.Printf("收到事件: %s, 类型: %s\n", evt.GetID(), evt.GetType())
        return nil
    }

    // 订阅事件
    sub.SubscribeWithCallback("user.created", myCallback)

    // 发布事件
    evt := event.NewBaseEvent("evt-001", "user.created", map[string]string{
        "username": "john_doe",
    })
    pub.Publish(evt)
}
```

运行基础示例:

```bash
go run examples/basic_example.go
```

运行高级示例:

```bash
go run examples/advanced_example.go
```

## 🏗️ 架构概述

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│  Publisher  │────────>│  Event Bus   │────────>│ Subscriber  │
└─────────────┘         │   (NATS)     │         └─────────────┘
                        └──────────────┘               │
                              │                        │
                              │                        v
                              │                 ┌─────────────┐
                              │                 │  Callbacks  │
                              │                 └─────────────┘
                              v
                        ┌──────────────┐
                        │   Events     │
                        └──────────────┘
```

## 🔧 核心组件

### 1. Event (事件)

事件是系统中传递的基本数据单元。

```go
type Event interface {
    GetID() string
    GetType() string
    GetTimestamp() time.Time
    GetPayload() interface{}
    Marshal() ([]byte, error)
}
```

### 2. EventBus (事件总线)

事件总线负责事件的发布和订阅，基于 NATS 实现。

```go
bus, err := eventbus.NewEventBus(eventbus.Config{
    NATSUrl: "nats://localhost:4222",
})
```

### 3. Publisher (发布者)

发布者用于向事件总线发布事件。

```go
pub := publisher.NewPublisher(bus)
pub.Publish(event)
```

### 4. Subscriber (订阅者)

订阅者用于订阅特定类型的事件。

```go
sub := subscriber.NewSubscriber(bus)
sub.SubscribeWithCallback("event.type", callbackFunc)
```

### 5. Callback (回调)

回调函数在事件到达时被执行。

```go
callback := func(ctx context.Context, evt event.Event) error {
    // 处理事件
    return nil
}
```

## 📚 使用示例

### 示例 1: 多个回调函数

```go
// 定义多个回调
callback1 := func(ctx context.Context, evt event.Event) error {
    fmt.Println("回调 1 执行")
    return nil
}

callback2 := func(ctx context.Context, evt event.Event) error {
    fmt.Println("回调 2 执行")
    return nil
}

// 注册回调
sub.SubscribeWithCallback("user.created", callback1)
sub.SubscribeWithCallback("user.created", callback2)

// 发布事件 - 两个回调都会被执行
pub.Publish(userEvent)
```

### 示例 2: 工作流回调链

```go
// 定义工作流步骤
callbacks := []callback.CallbackFunc{
    validateCallback,
    processCallback,
    notifyCallback,
}

// 批量注册回调
sub.SubscribeWithMultipleCallbacks("order.workflow", callbacks)
```

### 示例 3: 错误处理

```go
callback := func(ctx context.Context, evt event.Event) error {
    if err := processEvent(evt); err != nil {
        return fmt.Errorf("处理失败: %w", err)
    }
    return nil
}

sub.SubscribeWithCallback("event.type", callback)
```

## 📖 API 文档

### EventBus 方法

- `NewEventBus(config Config) (*EventBus, error)` - 创建新的事件总线
- `Publish(evt Event) error` - 发布事件
- `Subscribe(eventType string, handler func(Event) error) error` - 订阅事件
- `SubscribeWithCallback(eventType string, cb CallbackFunc) error` - 使用回调订阅
- `Unsubscribe(eventType string) error` - 取消订阅
- `Close()` - 关闭事件总线连接

### Publisher 方法

- `NewPublisher(eventBus *EventBus) *Publisher` - 创建发布者
- `Publish(evt Event) error` - 发布事件
- `PublishWithValidation(evt Event) error` - 带验证的发布

### Subscriber 方法

- `NewSubscriber(eventBus *EventBus) *Subscriber` - 创建订阅者
- `Subscribe(eventType string, handler func(Event) error) error` - 订阅事件
- `SubscribeWithCallback(eventType string, cb CallbackFunc) error` - 使用回调订阅
- `SubscribeMultiple(eventTypes []string, handler func(Event) error) error` - 订阅多个事件类型
- `SubscribeWithMultipleCallbacks(eventType string, callbacks []CallbackFunc) error` - 注册多个回调
- `Unsubscribe(eventType string) error` - 取消订阅

### Event 方法

- `NewBaseEvent(id, eventType string, payload interface{}) *BaseEvent` - 创建基础事件
- `GetID() string` - 获取事件 ID
- `GetType() string` - 获取事件类型
- `GetTimestamp() time.Time` - 获取时间戳
- `GetPayload() interface{}` - 获取负载
- `Marshal() ([]byte, error)` - 序列化为 JSON

## 🎯 使用场景

- **微服务通信**: 解耦微服务之间的依赖
- **异步处理**: 处理耗时任务而不阻塞主流程
- **事件溯源**: 记录和重放业务事件
- **实时通知**: 实时推送系统事件
- **工作流编排**: 构建复杂的业务流程

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📫 联系方式

- Email: 893400722@qq.com
- 兴趣: Go, JavaScript, Vue

## 📄 许可证

MIT License

---

由 @chengtb 开发维护
