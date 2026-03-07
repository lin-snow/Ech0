// package task declaration to use task related functionalities
package task

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/backup"
	"github.com/lin-snow/ech0/internal/event"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Tasker struct {
	scheduler      gocron.Scheduler
	commonService  *commonService.CommonService
	settingService *settingService.SettingService
	eventBus       event.IEventBus
	queueRepo      queueRepository.QueueRepositoryInterface
	started        bool
}

func (t *Tasker) Name() string {
	return "tasker"
}

func NewTasker(
	commonService *commonService.CommonService,
	settingService *settingService.SettingService,
	eventBusProvider func() event.IEventBus,
	queueRepo queueRepository.QueueRepositoryInterface,
) *Tasker {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logUtil.GetLogger().Error("Failed to create scheduler", zapcore.Field{
			Key:    "error",
			String: err.Error(),
		})
	}

	return &Tasker{
		scheduler:      scheduler,
		commonService:  commonService,
		settingService: settingService,
		eventBus:       eventBusProvider(),
		queueRepo:      queueRepo,
	}
}

func (t *Tasker) Start(context.Context) error {
	if t.started {
		return nil
	}
	if t.scheduler == nil {
		return errors.New("scheduler is nil")
	}
	if err := t.CleanupTempFilesTask(); err != nil {
		return err
	}
	if err := t.DeadLetterConsumeTask(); err != nil {
		return err
	}
	if err := t.InboxTask(); err != nil {
		return err
	}

	// 读取自动备份cron设置
	var backupScheduleSetting settingModel.BackupSchedule
	if err := t.settingService.GetBackupScheduleSetting(&backupScheduleSetting); err != nil {
		logUtil.GetLogger().
			Error("Failed to get backup schedule setting", zap.String("error", err.Error()))
		// 默认启用定时备份任务
		backupScheduleSetting.Enable = false
		backupScheduleSetting.CronExpression = "0 2 * * 0" // 每周日2点执行一次
	}
	if backupScheduleSetting.Enable {
		if err := t.ScheduleBackupTask(backupScheduleSetting.CronExpression); err != nil {
			return err
		}
	}

	t.scheduler.Start()
	t.started = true
	return nil
}

func (t *Tasker) Stop(context.Context) error {
	if !t.started || t.scheduler == nil {
		return nil
	}
	if err := t.scheduler.Shutdown(); err != nil {
		logUtil.GetLogger().Error("Failed to shutdown scheduler", zap.String("error", err.Error()))
		return err
	}
	t.started = false
	return nil
}

// CleanupTempFilesTask 清理过期的临时文件任务
func (t *Tasker) CleanupTempFilesTask() error {
	// 每三天执行一次
	_, err := t.scheduler.NewJob(
		gocron.DurationJob(72*time.Hour),
		gocron.NewTask(
			func() {
				if err := t.commonService.CleanupOrphanFiles(); err != nil {
					logUtil.GetLogger().
						Error("Failed to clean up temporary files", zap.String("error", err.Error()))
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule CleanupTempFilesTask", zap.String("error", err.Error()))
		return err
	}
	return nil
}

// DeadLetterConsumeTask 死信任务消费任务
func (t *Tasker) DeadLetterConsumeTask() error {
	// 每天12点执行一次, 测试时为每30秒执行一次
	_, err := t.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(12, 0, 0))),
		// gocron.DurationJob(30*time.Second), // 测试时为每30秒执行一次
		gocron.NewTask(
			func() {
				// 取出死信队列中的任务，逐个重试
				deadLetters, err := t.queueRepo.ListDeadLetters(context.Background(), 10)
				if err != nil {
					logUtil.GetLogger().
						Error("Failed To Get DeadLetters!", zap.String("error", err.Error()))
				}

				// 遍历死信任务，重新发送事件
				for _, dl := range deadLetters {
					// 发布事件到事件总线，触发重试
					if err := t.eventBus.Publish(context.Background(),
						event.NewEvent(
							event.EventTypeDeadLetterRetried,
							event.EventPayload{
								event.EventPayloadDeadLetter: dl,
							},
						),
					); err != nil {
						logUtil.GetLogger().
							Error("Failed to publish dead letter retried event", zap.String("error", err.Error()))
					}
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule WebhookRetryTask", zap.String("error", err.Error()))
		return err
	}
	return nil
}

// ScheduleBackupTask 定时备份任务
func (t *Tasker) ScheduleBackupTask(cronExpression string) error {
	// 判断 cron 表达式的字段数量来确定是否包含秒字段
	// 5 位表达式（分 时 日 月 周）：withSeconds = false
	// 6 位表达式（秒 分 时 日 月 周）：withSeconds = true
	withSeconds := false
	// 按空格分割 cron 表达式以准确判断字段数量
	fieldCount := len(strings.Fields(cronExpression))
	if fieldCount == 6 {
		withSeconds = true
	}

	_, err := t.scheduler.NewJob(
		gocron.CronJob(cronExpression, withSeconds),
		gocron.NewTask(
			func() {
				// 执行备份
				if path, fileName, err := backup.ExecuteBackup(); err != nil {
					logUtil.GetLogger().Error("Failed to execute scheduled backup",
						zap.String("path", path),
						zap.String("fileName", fileName),
						zap.String("error", err.Error()))
				}

				// 发布备份完成事件
				if err := t.eventBus.Publish(
					context.Background(),
					event.NewEvent(
						event.EventTypeSystemBackup,
						event.EventPayload{
							event.EventPayloadInfo: "System scheduled backup completed",
						},
					),
				); err != nil {
					logUtil.GetLogger().
						Error("Failed to publish backup completed event", zap.String("error", err.Error()))
				}
			},
		),
		gocron.WithTags("BackupSchedule"),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule ScheduleBackupTask", zap.String("error", err.Error()))
		return err
	}
	return nil
}

// InboxTask 定时处理Inbox任务
func (t *Tasker) InboxTask() error {
	// 每天12点执行一次, 测试时为每30秒执行一次
	_, err := t.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(12, 0, 0))),
		// gocron.DurationJob(30*time.Second), // 测试时为每30秒执行一次
		gocron.NewTask(
			func() {
				// 检查 Ech0 版本更新
				if err := t.eventBus.Publish(context.Background(),
					event.NewEvent(
						event.EventTypeEch0UpdateCheck,
						event.EventPayload{
							event.EventPayloadInfo: "Ech0 update checked",
						},
					),
				); err != nil {
					logUtil.GetLogger().Error("Failed to publish ech0 update checked event", zap.String("error", err.Error()))
				}

				// 清理已读的存在超过七天的消息
				if err := t.eventBus.Publish(context.Background(),
					event.NewEvent(
						event.EventTypeInboxClear,
						event.EventPayload{
							event.EventPayloadInfo: "Inbox cleared",
						},
					),
				); err != nil {
					logUtil.GetLogger().Error("Failed to publish inbox cleared event", zap.String("error", err.Error()))
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule InboxTask", zap.String("error", err.Error()))
		return err
	}
	return nil
}
