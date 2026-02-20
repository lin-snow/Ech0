package event

import (
	"context"
	"fmt"
	"time"

	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
)

type DeadLetterResolver struct {
	queueRepo queueRepository.QueueRepositoryInterface
	whd       *WebhookDispatcher
	fa        *FediverseAgent
}

func NewDeadLetterResolver(
	queueRepo queueRepository.QueueRepositoryInterface,
	whd *WebhookDispatcher,
	fa *FediverseAgent,
) *DeadLetterResolver {
	return &DeadLetterResolver{
		queueRepo: queueRepo,
		whd:       whd,
		fa:        fa,
	}
}

func (dlr *DeadLetterResolver) Handle(ctx context.Context, event *Event) error {
	// 取出 dead letter
	deadLetter, err := event.Payload[EventPayloadDeadLetter].(queueModel.DeadLetter)
	if !err {
		return fmt.Errorf("failed to extract dead letter: %v", err)
	}

	// 判断死信的状态
	switch deadLetter.Status {
	// 如果是待处理状态，尝试重新处理
	case queueModel.DeadLetterStatusPending:
		// 更新状态为处理中
		deadLetter.Status = queueModel.DeadLetterStatusProcessing
		deadLetter.RetryCount += 1
		deadLetter.UpdatedAt = time.Now().UTC()
		deadLetter.NextRetry = time.Now().UTC().Add(6 * time.Hour)

		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to processing: %v", err)
		}

		// 开始处理死信任务
		if err := dlr.processDeadLetter(ctx, &deadLetter); err != nil {
			// 处理失败，更新状态为失败
			deadLetter.ErrorMsg = err.Error()
			deadLetter.Status = queueModel.DeadLetterStatusFailed
			deadLetter.UpdatedAt = time.Now().UTC()
			if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
				return fmt.Errorf("failed to update dead letter to failed: %v", err)
			}
			return fmt.Errorf("failed to process dead letter: %v", err)
		}

		// 处理成功，更新状态为完成
		deadLetter.Status = queueModel.DeadLetterStatusCompleted
		deadLetter.UpdatedAt = time.Now().UTC()
		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter to completed: %v", err)
		}

		return nil

	// 处理中的死信任务，跳过
	case queueModel.DeadLetterStatusProcessing:
		// 处理中的死信任务，跳过
		return nil

	// 失败的死信任务，记录日志并跳过
	case queueModel.DeadLetterStatusFailed:
		// 失败的死信任务，检查重试次数
		if deadLetter.RetryCount <= 3 {
			// 重试次数未超过限制，更新状态为 pending 以便重新处理
			deadLetter.Status = queueModel.DeadLetterStatusPending
		} else {
			// 超过重试次数，更新状态为 discarded
			deadLetter.Status = queueModel.DeadLetterStatusDiscarded
		}

		// 更新死信任务
		deadLetter.UpdatedAt = time.Now().UTC()
		if err := dlr.queueRepo.UpdateDeadLetter(ctx, &deadLetter); err != nil {
			return fmt.Errorf("failed to update dead letter: %v", err)
		}
		return nil

	case queueModel.DeadLetterStatusDiscarded:
		// discarded 状态的死信任务，删除
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete discarded dead letter: %v", err)
		}
		return nil

	case queueModel.DeadLetterStatusCompleted:
		// completed 状态的死信任务，删除
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete completed dead letter: %v", err)
		}
		return nil

	default:
		// 未知状态，记录日志并修改状态为 discarded
		deadLetter.Status = queueModel.DeadLetterStatusDiscarded
		if err := dlr.queueRepo.DeleteDeadLetter(ctx, deadLetter.ID); err != nil {
			return fmt.Errorf("failed to delete unknown status dead letter: %v", err)
		}
		return fmt.Errorf("unknown dead letter status: %s", deadLetter.Status)
	}
}

// 处理死信任务
func (dlr *DeadLetterResolver) processDeadLetter(
	ctx context.Context,
	deadLetter *queueModel.DeadLetter,
) error {
	// 判断死信所属业务类型
	switch deadLetter.Type {
	case queueModel.DeadLetterTypeWebhook:
		// 处理 webhook 类型的死信任务
		return dlr.whd.HandleDeadLetter(ctx, deadLetter)

	case queueModel.DeadLetterTypePushEchoFediverse:
		// 处理 push echo federiverse 类型的死信任务
		return dlr.fa.HandlePushEchoDeadLetter(ctx, deadLetter)

	default:
		return fmt.Errorf("unknown dead letter type: %s", deadLetter.Type)
	}
}
