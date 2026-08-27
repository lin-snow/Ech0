# Busen

`Busen` 是一个小而清晰、typed-first、进程内的 Go 事件总线。

## 快速概览

- 定位：小而清晰、typed-first 的进程内事件总线
- 范围：只做单进程内分发，不做持久化、重放、跨进程投递
- API 风格：`Bus.Subscribe[T]` / `Bus.Publish[T]`，默认同步，语义直观
- 路由能力：支持精确 topic 与轻量通配（`*`、末尾 `>`）
- 并发控制：`Async()` + `WithBuffer(...)` + `WithOverflow(...)` 显式背压
- 顺序语义：支持 single-worker FIFO 与 per-subscriber/per-key 局部有序
- 可观测性：`Hooks` 观测 publish/error/panic/drop/reject
- 扩展点：`Use(...)` 中间件、`WithMetadata(...)` 元数据、`UseObserver(...)` 桥接观察

## 核心优势与能力

| 优势 | 价值 |
| --- | --- |
| typed-first API | `Bus.Subscribe[T]` / `Bus.Publish[T]` 直接用业务类型，减少样板代码和断言错误 |
| 显式并发语义 | sync/async、buffer、overflow、keyed ordering 都是可配置且可预期的 |
| 轻量但可扩展 | 支持 topic、middleware、hooks、metadata、observer，按需开启，不强制框架化 |
| 观测与排障友好 | 提供 publish/error/panic/drop/reject 生命周期回调，且携带结构化上下文 |
| 工程边界清晰 | 明确聚焦 in-process，不混入分布式消息平台职责，便于长期维护 |

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lin-snow/ech0/pkg/busen"
)

type UserCreated struct {
	ID    string
	Email string
}

func main() {
	bus := busen.New()

	unsubscribe, err := bus.Subscribe(func(ctx context.Context, event busen.Event[UserCreated]) error {
		fmt.Printf("welcome %s\n", event.Value.Email)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	err = bus.Publish(context.Background(), UserCreated{
		ID:    "u_123",
		Email: "hello@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	_ = bus.Close(context.Background())
}
```

## API 选择建议

大多数场景可以按下面方式选 API：

| 场景 | 建议 API |
| --- | --- |
| 只按类型收消息 | `Bus.Subscribe[T]` |
| 还需要按 topic 约束 | `Bus.SubscribeTopic[T]` |
| 一个 handler 需要订阅多个 topic | `Bus.SubscribeTopics[T]` |
| 需要按事件内容过滤 | `Bus.SubscribeMatch[T]` 或 `WithFilter(...)` |
| 希望调用方同步拿到 handler error | 默认同步订阅 |
| 希望异步投递并显式控制背压 | `Async()` + `WithBuffer(...)` + `WithOverflow(...)` |
| 希望单个订阅者 FIFO | `Sequential()` |
| 希望同一 key 局部有序 | `Async()` + `WithParallelism(...)` + 发布时 `WithKey(...)` |
| 希望观测 publish / panic / drop / reject | `WithHooks(...)` |
| 希望只包裹本地 handler 调用 | `Use(...)` 或 `WithMiddleware(...)` |
| 希望做 webhook/audit/落库桥接观察 | `UseObserver(...)` |

## 何时适合使用

| 适合使用 | 不适合使用 |
| --- | --- |
| 你希望在单个 Go 进程内做 typed event 解耦 | 你需要持久化、重放或跨进程投递 |
| 你需要轻量 topic 路由和有界异步投递 | 你需要内置 tracing、metrics、retry 或 rate limiting |
| 你希望保持 API 简洁并显式控制并发语义 | 你需要重型消息平台或分布式能力 |

## Topic 路由

`Busen` 支持点分隔的轻量 topic 路由。

- `*`：匹配恰好一个 segment
- `>`：匹配剩余的一个或多个 segment，且必须位于末尾

```go
sub, err := bus.SubscribeTopic("orders.>", func(ctx context.Context, event busen.Event[string]) error {
	fmt.Println(event.Topic, event.Value)
	return nil
})
if err != nil {
	log.Fatal(err)
}
defer sub()

_ = bus.Publish(context.Background(), "created", busen.WithTopic("orders.eu.created"))
```

如果同一个 handler 需要订阅多个 topic，也可以使用 `Bus.SubscribeTopics[T]`：

```go
sub, err := bus.SubscribeTopics([]string{"orders.created", "orders.updated"}, func(ctx context.Context, event busen.Event[string]) error {
	fmt.Println(event.Topic, event.Value)
	return nil
})
if err != nil {
	log.Fatal(err)
}
defer sub()

_ = bus.Publish(context.Background(), "created", busen.WithTopic("orders.created"))
_ = bus.Publish(context.Background(), "updated", busen.WithTopic("orders.updated"))
```

## 异步分发与顺序

异步订阅使用有界队列，背压策略是显式的：

- `OverflowBlock`
- `OverflowFailFast`
- `OverflowDropNewest`
- `OverflowDropOldest`

```go
_, err = bus.Subscribe(func(ctx context.Context, event busen.Event[UserCreated]) error {
	return nil
},
	busen.Async(),
	busen.Sequential(),
	busen.WithBuffer(128),
	busen.WithOverflow(busen.OverflowBlock),
)
```

如果发布时带上 `WithKey(...)`，那么同一 async 订阅者内、相同非空 ordering key 的事件会保持局部顺序：

```go
_, err = bus.Subscribe(func(ctx context.Context, event busen.Event[UserCreated]) error {
	return nil
}, busen.Async(), busen.WithParallelism(4), busen.WithBuffer(256))

_ = bus.Publish(context.Background(), UserCreated{ID: "1"}, busen.WithKey("tenant-a"))
_ = bus.Publish(context.Background(), UserCreated{ID: "2"}, busen.WithKey("tenant-a"))
```

边界说明：

- ordering key 只对 async subscriber 生效
- 空 key 会回退到普通非 keyed 调度
- 顺序保证是 **per subscriber / per key**，不是全局顺序

## Middleware 与 Hooks

### Middleware

`Busen` 提供一个很薄的 dispatch 中间件层，用来做本地组合，而不是重型 pipeline 框架。

```go
err = bus.Use(func(next busen.Next) busen.Next {
	return func(ctx context.Context, dispatch busen.Dispatch) error {
		return next(ctx, dispatch)
	}
})
if err != nil {
	log.Fatal(err)
}
```

中间件的边界：

- 只包 handler invocation
- 不替代钩子
- 不承担 retries、metrics、tracing、distributed concerns
- 不影响 subscriber matching、topic routing、async queue selection
- 对 `Dispatch` 的修改只影响后续中间件和最终 handler
- 钩子仍然观察原始 publish 元数据

如果你更喜欢构造期注册方式，也可以使用 `WithMiddleware(...)`：

```go
bus := busen.New(
	busen.WithMiddleware(func(next busen.Next) busen.Next {
		return func(ctx context.Context, dispatch busen.Dispatch) error {
			return next(ctx, dispatch)
		}
	}),
)
```

### Hooks

`Hooks` 用来观察运行时事件，而不是参与 handler 调用链控制。

```go
bus := busen.New(
	busen.WithHooks(busen.Hooks{
		OnPublishDone: func(info busen.PublishDone) {
			log.Printf("matched=%d delivered=%d err=%v", info.MatchedSubscribers, info.DeliveredSubscribers, info.Err)
		},
		OnHandlerError: func(info busen.HandlerError) {
			log.Printf("async=%v topic=%q err=%v", info.Async, info.Topic, info.Err)
		},
		OnHandlerPanic: func(info busen.HandlerPanic) {
			log.Printf("panic in %v: %v", info.EventType, info.Value)
		},
		OnEventDropped: func(info busen.DroppedEvent) {
			log.Printf("dropped event for topic %q with policy %v", info.Topic, info.Policy)
		},
		OnEventRejected: func(info busen.RejectedEvent) {
			log.Printf("rejected event for topic %q with policy %v", info.Topic, info.Policy)
		},
	}),
)
```

当前 hooks 触发点包括：

- `OnPublishStart`
- `OnPublishDone`
- `OnHandlerError`
- `OnHandlerPanic`
- `OnEventDropped`
- `OnEventRejected`
- `OnHookPanic`

### 可选统一 metadata

`Busen` 保持 typed-first，不强制事件信封；如果你需要统一结构化元数据，可以按需启用 metadata 层。

```go
bus := busen.New(
	busen.WithMetadataBuilder(func(input busen.PublishMetadataInput) map[string]string {
		return map[string]string{
			"source": "billing-service",
		}
	}),
)

_ = bus.Publish(
	context.Background(),
	OrderCreated{ID: "o_1"},
	busen.WithMetadata(map[string]string{
		"trace_id": "tr_123",
	}),
)
```

规则：

- builder 是全局默认 metadata
- `WithMetadata(...)` 的同名键会覆盖 builder 返回值
- metadata 会传递到 middleware、handler 以及 hooks

### Observer（桥接观察者）

`UseObserver(...)` 用于横切观察（webhook、审计、事件落库等），语义是“观察已接受投递”，不参与主业务订阅匹配。

```go
_ = bus.UseObserver(
	func(ctx context.Context, obs busen.Observation) {
		log.Printf("observe type=%v topic=%q sub=%d", obs.EventType, obs.Topic, obs.SubscriberID)
	},
	busen.ObserveTopic("orders.>"),
	busen.ObserveMetadata(map[string]string{"trace_id": "tr_123"}),
)
```

可选过滤器：

- `ObserveType[T]()`
- `ObserveTopic(pattern)`
- `ObserveMetadata(map[string]string)`
- `ObserveMatch(func(Observation) bool)`

### Shutdown 模式

`Close(ctx)` 保持兼容，等价于 `Shutdown(ctx, ShutdownDrain)`。如果你需要更细粒度策略，可以使用 `Shutdown(...)`：

```go
result, err := bus.Shutdown(context.Background(), busen.ShutdownBestEffort)
if err != nil {
	log.Fatal(err)
}
log.Printf("completed=%v processed=%d dropped=%d rejected=%d timeout_subs=%v",
	result.Completed, result.Processed, result.Dropped, result.Rejected, result.TimedOutSubscribers)
```

模式说明：

- `ShutdownDrain`：停止接收并尽量完整 drain（`Close` 默认语义）
- `ShutdownBestEffort`：停止接收后尽力等待到 `ctx` 截止
- `ShutdownAbort`：停止接收并立即丢弃队列中待处理事件（不强制终止正在运行的 handler）

## 性能测试

`Busen` 内置了可重复运行的 benchmark，覆盖 `Bus.Publish[T]`、sync/async、topic 路由、middleware、hooks 等热路径。

主要覆盖项：

- `Bus.Publish[T]` 在 `1 / 10 / 100` 个订阅者下的成本
- sync 与 async sequential 的差异
- async keyed delivery
- middleware 开启/关闭
- middleware + hooks 同时开启
- async keyed + topic routing
- exact / wildcard 路由
- direct router matcher 成本

运行方式：

```bash
go test ./... -run '^$' -bench . -benchmem
```

这些数字代表的是 **in-process event bus 的热路径开销**，不是消息系统吞吐保证。

在一台使用 Go `1.27.0` 的 Apple M4 机器上，`-count=5` 取中位数的参考结果大致为：

| 场景 | 参考耗时 |
| --- | --- |
| sync publish（1 subscriber） | 约 `123 ns/op` |
| sync publish（10 subscribers） | 约 `658 ns/op` |
| async sequential publish | 约 `204 ns/op` |
| async keyed publish | 约 `282 ns/op` |
| middleware-enabled publish | 约 `128 ns/op` |
| middleware + hooks publish | 约 `145 ns/op` |
| async keyed + topic publish | 约 `270 ns/op` |
| exact topic publish | 约 `138 ns/op` |
| wildcard topic publish | 约 `143 ns/op` |

这一轮里，router matcher 依然保持 `0 allocs/op`：

| matcher | 参考耗时 | 分配 |
| --- | --- | --- |
| exact matcher | 约 `1.8 ns/op` | `0 allocs/op` |
| wildcard matcher | 约 `5.9 ns/op` | `0 allocs/op` |

新增能力（metadata / observer）的一轮参考结果如下：

| 场景 | 参考耗时 | 分配 |
| --- | --- | --- |
| publish with metadata（disabled） | 约 `122 ns/op` | `288 B/op`, `4 allocs/op` |
| publish with metadata（enabled） | 约 `787 ns/op` | `2640 B/op`, `18 allocs/op` |
| publish with observer（disabled） | 约 `137 ns/op` | `312 B/op`, `5 allocs/op` |
| publish with observer（enabled） | 约 `174 ns/op` | `376 B/op`, `6 allocs/op` |

说明：

- 上表来自 `go test ./... -run '^$' -bench . -benchmem -count=5` 的中位数样本，主要用于量级感知
- `metadata` 开销主要来自 map 复制/合并与 hook/handler 透传
- `observer` 在“仅观察、轻过滤”下增量较小；复杂过滤函数会抬高开销
- 建议在你的目标硬件上用相同命令复测后再做容量预算

## 设计边界

- 类型匹配是精确匹配，不做接口层级自动分发
- 不保证全局顺序，只保证局部顺序语义
- sync handler 错误会直接返回给 `Publish`
- async handler error / panic 不回传给 `Publish`，应通过 `Hooks` 观测
- Busen 保证的是分发链路并发安全，不保证 `event.Value`（`any`）内部可变对象的线程安全；建议发布后视为不可变，或由业务自行拷贝/加锁
- `Close(ctx)` 超时表示未在期限内 drain 完成，不会强制终止用户 handler
- 这是 in-process event bus，不是 distributed event platform

## 相关文档

- 更多用法示例：`example_test.go`
