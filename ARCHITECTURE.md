# EDA Architecture Documentation

## 项目结构

```
chengtb/
├── pkg/
│   └── eda/
│       ├── event/          # 事件定义
│       │   └── event.go
│       ├── callback/       # 回调处理
│       │   └── callback.go
│       ├── eventbus/       # NATS事件总线
│       │   └── eventbus.go
│       ├── publisher/      # 事件发布者
│       │   └── publisher.go
│       └── subscriber/     # 事件订阅者
│           └── subscriber.go
├── examples/
│   ├── basic/             # 基础示例
│   │   └── main.go
│   └── advanced/          # 高级示例
│       └── main.go
├── go.mod
├── go.sum
└── README.md
```

## 核心概念

### 1. 事件 (Event)

事件是 EDA 系统中的核心概念，代表系统中发生的一个动作或状态变化。

**特性:**
- 唯一 ID 标识
- 事件类型分类
- 时间戳记录
- 灵活的负载数据
- JSON 序列化支持

**示例:**
```go
event := event.NewBaseEvent(
    "evt-001",              // 事件ID
    "user.created",         // 事件类型
    UserCreatedPayload{     // 负载
        UserID: "user-123",
        Username: "john",
    },
)
```

### 2. 事件总线 (EventBus)

事件总线是事件分发的中枢，基于 NATS 消息中间件实现。

**职责:**
- 连接 NATS 服务器
- 发布事件到相应主题
- 订阅事件类型
- 管理订阅关系
- 触发回调执行

**NATS 主题命名:**
- 格式: `events.{eventType}`
- 示例: `events.user.created`、`events.order.placed`

### 3. 发布者 (Publisher)

发布者提供便捷的事件发布接口。

**功能:**
- 发布事件
- 事件验证
- 错误处理

### 4. 订阅者 (Subscriber)

订阅者管理事件订阅和回调注册。

**功能:**
- 订阅单个/多个事件类型
- 注册回调函数
- 批量回调注册
- 取消订阅

### 5. 回调 (Callback)

回调函数在事件到达时被异步执行。

**特性:**
- 并发执行多个回调
- 上下文传递
- 错误收集
- 同步/异步执行模式

## 工作流程

### 事件发布流程

```
1. Publisher.Publish(event)
   ↓
2. EventBus.Publish(event)
   ↓
3. event.Marshal() → JSON
   ↓
4. NATS.Publish("events.{type}", data)
   ↓
5. NATS 分发到订阅者
```

### 事件订阅流程

```
1. Subscriber.SubscribeWithCallback(eventType, callback)
   ↓
2. CallbackHandler.Register(eventType, callback)
   ↓
3. EventBus.SubscribeWithCallback(eventType, callback)
   ↓
4. NATS.Subscribe("events.{type}", handler)
   ↓
5. 接收到事件时:
   a. Unmarshal JSON → Event
   b. CallbackHandler.Execute(event)
   c. 异步执行所有注册的回调
```

## 回调执行机制

### 异步执行 (Execute)

```go
func (h *CallbackHandler) Execute(ctx context.Context, evt event.Event) error
```

**特点:**
- 所有回调并发执行
- 使用 Goroutine 和 WaitGroup
- 收集所有错误
- 不阻塞主流程

**流程:**
```
Event 到达
  ↓
启动多个 Goroutine (每个回调一个)
  ↓
并发执行所有回调
  ↓
等待所有完成 (WaitGroup)
  ↓
收集错误并返回
```

### 同步执行 (ExecuteSync)

```go
func (h *CallbackHandler) ExecuteSync(ctx context.Context, evt event.Event) error
```

**特点:**
- 按注册顺序执行
- 遇到错误立即返回
- 适合需要保证顺序的场景

## 使用模式

### 模式 1: 简单发布订阅

```go
// 订阅
sub.SubscribeWithCallback("event.type", func(ctx context.Context, evt event.Event) error {
    // 处理事件
    return nil
})

// 发布
pub.Publish(event)
```

### 模式 2: 多回调处理

```go
// 为同一事件注册多个处理器
sub.SubscribeWithCallback("user.created", sendEmailCallback)
sub.SubscribeWithCallback("user.created", updateAnalyticsCallback)
sub.SubscribeWithCallback("user.created", sendNotificationCallback)

// 发布一次，三个回调都会执行
pub.Publish(userCreatedEvent)
```

### 模式 3: 工作流编排

```go
// 定义工作流步骤
callbacks := []callback.CallbackFunc{
    validateOrder,
    processPayment,
    updateInventory,
    sendConfirmation,
}

// 批量注册
sub.SubscribeWithMultipleCallbacks("order.workflow", callbacks)
```

### 模式 4: 错误处理

```go
callback := func(ctx context.Context, evt event.Event) error {
    // 业务逻辑
    if err := doSomething(); err != nil {
        // 记录错误
        log.Printf("Error: %v", err)
        return err
    }
    return nil
}
```

## 最佳实践

### 1. 事件设计

- **明确的事件类型**: 使用清晰的命名，如 `user.created`、`order.placed`
- **完整的负载**: 包含足够的信息，避免额外查询
- **不可变性**: 事件一旦发布，不应修改
- **版本控制**: 考虑事件模式的演进

### 2. 回调实现

- **幂等性**: 回调应该是幂等的，可以安全地重复执行
- **快速返回**: 避免长时间阻塞，考虑异步处理
- **错误处理**: 妥善处理错误，不影响其他回调
- **超时控制**: 使用 context 进行超时控制

### 3. 性能优化

- **连接复用**: 复用 EventBus 连接
- **批量操作**: 批量注册回调
- **资源清理**: 及时关闭连接和取消订阅
- **并发控制**: 控制回调并发数

### 4. 监控和日志

- **事件追踪**: 记录事件ID和处理流程
- **错误监控**: 监控回调执行错误
- **性能指标**: 监控事件处理延迟
- **审计日志**: 记录关键业务事件

## 故障处理

### NATS 连接失败

```go
bus, err := eventbus.NewEventBus(config)
if err != nil {
    // 重试逻辑
    // 降级处理
    // 报警通知
}
```

### 回调执行失败

```go
callback := func(ctx context.Context, evt event.Event) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Callback panic: %v", r)
        }
    }()
    // 业务逻辑
}
```

### 消息堆积

- 增加订阅者数量
- 优化回调执行效率
- 使用队列分组
- 监控消息延迟

## 扩展性

### 水平扩展

- NATS 支持集群模式
- 多个订阅者实例共享负载
- 使用队列组 (Queue Groups)

### 功能扩展

- **事件过滤**: 基于内容的过滤
- **事件转换**: 事件格式转换
- **死信队列**: 处理失败事件
- **事件重放**: 历史事件回放
- **事件存储**: 事件持久化

## 安全考虑

- **认证**: NATS 连接认证
- **授权**: 基于主题的权限控制
- **加密**: TLS 加密传输
- **审计**: 事件访问审计
- **限流**: 防止事件洪泛

## 测试策略

### 单元测试

```go
// 测试事件创建
func TestNewBaseEvent(t *testing.T) {
    evt := event.NewBaseEvent("id", "type", payload)
    // 断言...
}

// 测试回调执行
func TestCallbackExecution(t *testing.T) {
    handler := callback.NewCallbackHandler()
    // 测试...
}
```

### 集成测试

```go
// 测试完整流程
func TestEventPublishSubscribe(t *testing.T) {
    // 启动 NATS
    // 创建 EventBus
    // 发布订阅测试
}
```

### 性能测试

```go
// 基准测试
func BenchmarkEventPublish(b *testing.B) {
    for i := 0; i < b.N; i++ {
        pub.Publish(event)
    }
}
```

## 总结

这个 EDA 架构提供了:

1. ✅ **解耦**: 发布者和订阅者解耦
2. ✅ **可扩展**: 易于添加新的事件处理器
3. ✅ **高性能**: 基于 NATS 的高性能消息传输
4. ✅ **灵活性**: 支持多种使用模式
5. ✅ **可靠性**: 完善的错误处理机制

适用于构建现代微服务架构和事件驱动系统。
