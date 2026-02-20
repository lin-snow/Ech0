package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lin-snow/ech0/internal/async"
	"github.com/lin-snow/ech0/internal/fediverse"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	"github.com/lin-snow/ech0/internal/transaction"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

type PushEchoReplayPayload struct {
	Echo echoModel.Echo `json:"echo"`
	User userModel.User `json:"user"`
}

// FediverseAgent 处理联邦相关的事件
type FediverseAgent struct {
	pool      *async.WorkerPool                        // 任务池
	core      *fediverse.FediverseCore                 // 联邦宇宙相关组件
	queueRepo queueRepository.QueueRepositoryInterface // 死信任务仓储
	txManager transaction.TransactionManager           // 事务管理器
}

func NewFediverseAgent(
	core *fediverse.FediverseCore,
	queueRepo queueRepository.QueueRepositoryInterface,
	txManager transaction.TransactionManager,
) *FediverseAgent {
	return &FediverseAgent{
		pool:      async.NewWorkerPool(3, 3), // 假设最大并发数为 3，任务队列大小为 3
		core:      core,
		queueRepo: queueRepo,
		txManager: txManager,
	}
}

func (fa *FediverseAgent) Handle(ctx context.Context, e *Event) error {
	// 处理事件，与联邦宇宙交互
	switch e.Type {
	case EventTypeEchoCreated:
		if err := fa.HandleCreateEchoEvent(ctx, e); err != nil {
			logUtil.GetLogger().
				Error("Failed to handle create echo event", zap.String("error", err.Error()))
			return nil
		}

	default:
		return nil // 忽略其他事件
	}

	return nil
}

func (fa *FediverseAgent) Wait() {
	fa.pool.Wait()
}

func (fa *FediverseAgent) HandleCreateEchoEvent(ctx context.Context, e *Event) error {
	// 将 Echo 推送到联邦宇宙
	payload := e.Payload
	echoData, ok := payload[EventPayloadEcho]
	if !ok {
		return nil
	}
	echo, ok := echoData.(echoModel.Echo)
	if !ok {
		return nil
	}
	userData, ok := payload[EventPayloadUser]
	if !ok {
		return nil
	}
	user, ok := userData.(userModel.User)
	if !ok {
		return nil
	}

	fa.pool.Submit(func() error {
		// 重试机制，最多重试3次，初始延迟1秒
		return fa.retryWithBackoff(3, time.Second, func() error {
			if err := fa.core.PushEchoToFediverse(user.ID, echo); err != nil {
				logUtil.GetLogger().Error(err.Error())

				// 处理失败，记录到死信队列
				e.Meta = map[string]any{
					queueModel.DeadLetterMetaKey: true, // 标记为死信任务
				}

				payloadData := PushEchoReplayPayload{
					Echo: echo,
					User: user,
				}
				payload, _ := json.Marshal(payloadData)

				// 保存到死信队列
				var deadLetter queueModel.DeadLetter
				deadLetter.SetType(queueModel.DeadLetterTypePushEchoFediverse)
				deadLetter.Payload = payload
				deadLetter.ErrorMsg = err.Error()
				deadLetter.RetryCount = 0
				deadLetter.NextRetry = time.Now().UTC().Add(12 * time.Hour) // 初始重试时间为 12 小时后
				deadLetter.CreatedAt = time.Now().UTC()
				deadLetter.UpdatedAt = time.Now().UTC()
				deadLetter.Status = queueModel.DeadLetterStatusPending // 初始状态为待处理

				if err := fa.txManager.Run(func(ctx context.Context) error {
					return fa.queueRepo.SaveDeadLetter(ctx, &deadLetter)
				}); err != nil {
					logUtil.GetLogger().
						Error("Failed to save dead letter", zap.String("error", err.Error()))
				}
			}
			return nil
		})
	})

	return nil
}

func (fa *FediverseAgent) retryWithBackoff(
	retries int,
	delay time.Duration,
	fn func() error,
) error {
	var err error
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}

func (fa *FediverseAgent) HandlePushEchoDeadLetter(
	ctx context.Context,
	deadLetter *queueModel.DeadLetter,
) error {
	// 解析负载
	var payload PushEchoReplayPayload
	if err := json.Unmarshal(deadLetter.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal dead letter payload: %w", err)
	}
	echo := payload.Echo
	user := payload.User

	// 重试
	err := fa.retryWithBackoff(3, 1*time.Minute, func() error {
		return fa.core.PushEchoToFediverse(user.ID, echo)
	})
	if err != nil {
		return err
	}

	return nil
}
