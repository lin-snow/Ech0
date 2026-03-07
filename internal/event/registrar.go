package event

import "sync/atomic"

// EventHandlers 事件处理器集合
type EventHandlers struct {
	wbd *WebhookDispatcher  // webhook 事件处理器
	dlr *DeadLetterResolver // 死信处理器
	bs  *BackupScheduler    // 备份事件调度器
	ap  *AgentProcessor     // Agent事件处理器
	id  *InboxDispatcher    // Inbox事件处理器
}

// NewEventHandlers 创建一个新的事件处理器集合
func NewEventHandlers(
	wbd *WebhookDispatcher,
	dlr *DeadLetterResolver,
	bs *BackupScheduler,
	ap *AgentProcessor,
	id *InboxDispatcher,
) *EventHandlers {
	return &EventHandlers{wbd: wbd, dlr: dlr, bs: bs, ap: ap, id: id}
}

// EventRegistrar 事件注册器
type EventRegistrar struct {
	eb         IEventBus      // 事件总线
	eh         *EventHandlers // 事件处理器集合
	registered atomic.Bool
}

// NewEventRegistry 创建一个新的事件注册表
func NewEventRegistry(ebp func() IEventBus, eh *EventHandlers) *EventRegistrar {
	return &EventRegistrar{eb: ebp(), eh: eh}
}

// Register 注册事件处理函数
func (er *EventRegistrar) Register() error {
	if er.registered.Load() {
		return nil
	}

	var err error
	// 订阅死信事件
	err = er.eb.Subscribe(
		er.eh.dlr.Handle,
		EventTypeDeadLetterRetried,
	) // 订阅死信事件，交给 DeadLetterResolver 处理
	if err != nil {
		return err
	}
	err = er.eb.Subscribe(
		er.eh.bs.Handle,
		EventTypeUpdateBackupSchedule,
	) // 订阅 UpdateBackupSchedule 事件，交给 BackupScheduler 处理
	if err != nil {
		return err
	}

	// 订阅事件组
	err = er.eb.Subscribes(
		er.eh.ap.Handle,
		EventTypeEchoCreated,
		EventTypeUserDeleted,
		EventTypeEchoUpdated,
	) // 订阅 Echo 事件组，交给 AgentProcessor 处理
	if err != nil {
		return err
	}

	// 订阅 Inbox 事件，交给 InboxDispatcher 处理
	err = er.eb.Subscribes(
		er.eh.id.Handle,
		EventTypeEch0UpdateCheck,
		EventTypeInboxClear,
	) // 订阅 Inbox 事件，交给 InboxDispatcher 处理
	if err != nil {
		return err
	}

	// 订阅所有事件，交给 WebhookDispatcher 处理
	err = er.eb.SubscribeAll(
		er.eh.wbd.Handle,
		EventTypeDeadLetterRetried,
	) // 订阅所有事件，交给 WebhookDispatcher 处理,但是排除死信事件
	if err != nil {
		return err
	}

	er.registered.Store(true)
	return err
}

// Stop 等待已投递的异步处理任务完成。
func (er *EventRegistrar) Stop() error {
	if !er.registered.Load() {
		return nil
	}
	er.eh.wbd.Wait()
	return nil
}
