package event

import (
	"context"
	"math/rand"
	"sync"
	"time"

	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

// 定义事件类型
const (
	EventTypeUserCreated EventType = "user.created" // 创建用户
	EventTypeUserUpdated EventType = "user.updated" // 更新用户
	EventTypeUserDeleted EventType = "user.deleted" // 删除用户

	EventTypeEchoCreated EventType = "echo.created" // 创建Echo
	EventTypeEchoUpdated EventType = "echo.updated" // 更新Echo
	EventTypeEchoDeleted EventType = "echo.deleted" // 删除Echo

	EventTypeResourceUploaded EventType = "resource.uploaded" // 资源上传

	EventTypeSystemBackup         EventType = "system.backup"                 // 系统快照备份
	EventTypeSystemRestore        EventType = "system.restore"                // 系统快照恢复
	EventTypeSystemExport         EventType = "system.export"                 // 系统快照导出
	EventTypeUpdateBackupSchedule EventType = "system.update_backup_schedule" // 更新自动备份计划

	EventTypeDeadLetterRetried EventType = "deadletter.retried" // 死信任务重试

	EventTypeInboxClear EventType = "inbox.clear" // 清理Inbox（超过七天的已读消息）

	EventTypeEch0UpdateCheck EventType = "ech0.update" // 检查 Ech0 版本更新
)

// 定义事件Payload的常用字段
const (
	EventPayloadUser       = "user"
	EventPayloadEcho       = "echo"
	EventPayloadData       = "data"
	EventPayloadSchedule   = "schedule"
	EventPayloadInfo       = "info"
	EventPayloadMeta       = "meta"
	EventPayloadTime       = "time"
	EventPayloadType       = "type"
	EventPayloadID         = "id"
	EventPayloadURL        = "url"
	EventPayloadSize       = "size"
	EventPayloadPath       = "path"
	EventPayloadFile       = "file"
	EventPayloadDeadLetter = "dead_letter"
)

// Event 事件结构体
type Event struct {
	ID        string         `json:"id"`        // Unique event ID
	Type      EventType      `json:"type"`      // Event type
	Payload   EventPayload   `json:"data"`      // Event payload
	Timestamp time.Time      `json:"timestamp"` // Event timestamp
	Meta      map[string]any `json:"meta"`      // 事件元数据
}

type (
	EventType    string         // 事件类型
	EventPayload map[string]any // 事件负载
)

// NewEvent 创建一个新的事件
func NewEvent(eventType EventType, payload EventPayload, meta ...map[string]any) *Event {
	now := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(now), rand.New(rand.NewSource(now.UnixNano()))).
		String()
		// 使用 ULID 生成唯一 ID
	return &Event{
		ID:        id,
		Type:      eventType,
		Payload:   payload,
		Timestamp: now,
		Meta: func() map[string]any {
			if len(meta) > 0 {
				return meta[0]
			}
			return nil
		}(),
	}
}

// IEventBus 事件总线接口
type IEventBus interface {
	Publish(ctx context.Context, event *Event) error                // 发布事件
	Subscribe(handler EventHandler, eventType EventType) error      // 订阅特定事件
	Subscribes(handler EventHandler, eventTypes ...EventType) error // 订阅多个特定事件
	SubscribeAll(handler EventHandler, exclude ...EventType) error  // 订阅所有事件
}

// EventHandler 事件处理函数类型
type EventHandler func(ctx context.Context, event *Event) error

// EventFilter 事件过滤器函数类型
type EventFilter func(eventType EventType) bool

// globalHandler 全局事件处理器
type globalHandler struct {
	handler EventHandler
	filter  EventFilter
}

// EventBus 事件总线
type EventBus struct {
	mu   sync.RWMutex                 // 读写锁保护订阅者列表
	subs map[EventType][]EventHandler // 订阅者列表
	all  []globalHandler              // 支持过滤的全局订阅者
}

// NewEventBus 创建一个新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[EventType][]EventHandler),
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	eb.mu.RLock()
	handlers, ok := eb.subs[event.Type]
	eb.mu.RUnlock()
	if !ok {
		handlers = []EventHandler{}
	}

	// 处理所有订阅该事件类型的处理器
	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
				// 错误处理
				logUtil.GetLogger().Error("Event Handler Error:", zap.String("err", err.Error()))
				// log.Println("Event Handler Error:", err)
			}
		}(handler)
	}

	eb.mu.RLock()
	allHandlers := eb.all
	eb.mu.RUnlock()

	// 处理所有订阅所有事件的处理器
	for _, gh := range allHandlers {
		if gh.filter == nil || gh.filter(event.Type) {
			go func(h EventHandler) {
				if err := h(ctx, event); err != nil {
					// 错误处理
					logUtil.GetLogger().
						Error("Event Handler Error:", zap.String("err", err.Error()))

					// log.Println("Event Handler Error:", err)
				}
			}(gh.handler)
		}
	}

	return nil
}

// Subscribe 订阅指定事件
func (eb *EventBus) Subscribe(handler EventHandler, eventType EventType) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subs[eventType] = append(eb.subs[eventType], handler)
	return nil
}

// Subscribes 订阅多个指定事件
func (eb *EventBus) Subscribes(handler EventHandler, eventTypes ...EventType) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, et := range eventTypes {
		eb.subs[et] = append(eb.subs[et], handler)
	}
	return nil
}

// SubscribeAll 订阅所有事件
func (eb *EventBus) SubscribeAll(handler EventHandler, exclude ...EventType) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	excludeSet := make(map[EventType]struct{}, len(exclude))
	for _, et := range exclude {
		excludeSet[et] = struct{}{}
	}

	filter := func(eventType EventType) bool {
		_, skip := excludeSet[eventType] // 如果在排除列表中,则skip为true
		return !skip
	}

	eb.all = append(eb.all, globalHandler{
		handler: handler,
		filter:  filter,
	})
	return nil
}
