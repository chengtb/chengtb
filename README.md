- 👋 Hi, I'm @chengtb
- 👀 I'm interested in Go Javascript Vue
- 🌱 I'm currently learning [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- 💞️ I'm looking to collaborate on ...
- 📫 How to reach me 893400722@qq.com
- 😄 Pronouns: ...
- ⚡ Fun fact: ...

---

## 📖 Reading Notes: [temporalio/sdk-go](https://github.com/temporalio/sdk-go)

[Temporal](https://github.com/temporalio/temporal) 是一个分布式、可扩展、持久且高可用的编排引擎，用于以可扩展和弹性的方式执行异步长时间运行的业务逻辑。

**Temporal Go SDK** 是 Temporal 用于在 Go 语言中编写 Workflow 和 Activity 的框架。

### 核心概念

| 概念 | 说明 |
|------|------|
| **Workflow** | 描述业务流程的代码，必须是确定性的（deterministic） |
| **Activity** | 执行实际副作用的代码（如 HTTP 请求、数据库操作） |
| **Worker** | 运行 Workflow 和 Activity 的进程 |
| **Task Queue** | Workflow/Activity 任务的分发队列 |
| **Client** | 与 Temporal 服务通信，用于启动 Workflow 等操作 |

### 快速上手

```bash
git clone https://github.com/temporalio/sdk-go.git
```

```go
package main

import (
    "context"
    "time"

    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
    "go.temporal.io/sdk/workflow"
)

// Workflow 定义
func MyWorkflow(ctx workflow.Context, name string) (string, error) {
    ao := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
    ctx = workflow.WithActivityOptions(ctx, ao)
    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity, name).Get(ctx, &result)
    return result, err
}

// Activity 定义
func MyActivity(ctx context.Context, name string) (string, error) {
    return "Hello, " + name + "!", nil
}
```

### 关键包

- [`go.temporal.io/sdk/client`](https://pkg.go.dev/go.temporal.io/sdk/client) — 创建 Temporal 客户端，启动/查询 Workflow
- [`go.temporal.io/sdk/worker`](https://pkg.go.dev/go.temporal.io/sdk/worker) — 创建并运行 Worker
- [`go.temporal.io/sdk/workflow`](https://pkg.go.dev/go.temporal.io/sdk/workflow) — 编写 Workflow 逻辑
- [`go.temporal.io/sdk/activity`](https://pkg.go.dev/go.temporal.io/sdk/activity) — 编写 Activity 逻辑
- [`go.temporal.io/sdk/converter`](https://pkg.go.dev/go.temporal.io/sdk/converter) — 数据序列化转换
- [`go.temporal.io/sdk/temporal`](https://pkg.go.dev/go.temporal.io/sdk/temporal) — 错误类型、重试策略等

### Workflow 确定性要求

Workflow 代码必须是**确定性**的，不能直接使用：
- `time.Now()` → 使用 `workflow.Now(ctx)`
- `time.Sleep()` → 使用 `workflow.Sleep(ctx, duration)`
- `math/rand` 随机数 → 使用 `workflow.SideEffect`
- 直接发起网络请求 → 改写为 Activity

可以使用官方工具检测不确定性：[contrib/tools/workflowcheck](https://github.com/temporalio/sdk-go/blob/master/contrib/tools/workflowcheck)

### 使用 slog（Go 1.21+）

```go
clientOptions := client.Options{
    Logger: log.NewStructuredLogger(
        slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            AddSource: true,
            Level:     slog.LevelDebug,
        }))),
}
temporalClient, err := client.Dial(clientOptions)
```

### 参考资料

- 📚 [官方文档](https://docs.temporal.io)
- 📦 [API 文档](https://pkg.go.dev/go.temporal.io/sdk)
- 🧪 [示例代码](https://github.com/temporalio/samples-go)
- 🔧 [源码仓库](https://github.com/temporalio/sdk-go)

<!---
chengtb/chengtb is a ✨ special ✨ repository because its `README.md` (this file) appears on your GitHub profile.
You can click the Preview link to take a look at your changes.
--->
