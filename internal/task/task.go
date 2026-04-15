// package task declaration to use task related functionalities
package task

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/backup"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"github.com/lin-snow/ech0/internal/visitor"
	"go.uber.org/zap"
)

type Tasker struct {
	scheduler      gocron.Scheduler
	fileService    fileService.Service
	settingService settingService.Service
	publisher      *publisher.Publisher
	queueRepo      *queueRepository.QueueRepository
	storageManager *storage.Manager
	visitorTracker *visitor.Tracker
	visitorRepo    *visitorRepository.VisitorRepository
	started        bool
	mu             sync.Mutex
}

const (
	backupScheduleTag  = "BackupSchedule"
	visitorSnapshotTag = "VisitorSnapshotSchedule"
)

func (t *Tasker) Name() string {
	return "tasker"
}

func NewTasker(
	fileSvc fileService.Service,
	settingService settingService.Service,
	publisher *publisher.Publisher,
	queueRepo *queueRepository.QueueRepository,
	storageManager *storage.Manager,
	visitorTracker *visitor.Tracker,
	visitorRepo *visitorRepository.VisitorRepository,
) *Tasker {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logUtil.GetLogger().Error("Failed to create scheduler", zap.Error(err))
	}

	return &Tasker{
		scheduler:      scheduler,
		fileService:    fileSvc,
		settingService: settingService,
		publisher:      publisher,
		queueRepo:      queueRepo,
		storageManager: storageManager,
		visitorTracker: visitorTracker,
		visitorRepo:    visitorRepo,
	}
}

func (t *Tasker) Start(context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

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
	// 读取自动备份cron设置
	var backupScheduleSetting settingModel.BackupSchedule
	if err := t.settingService.GetBackupScheduleSetting(&backupScheduleSetting); err != nil {
		logUtil.GetLogger().
			Error("Failed to get backup schedule setting", zap.Error(err))
		// 默认启用定时备份任务
		backupScheduleSetting.Enable = false
		backupScheduleSetting.CronExpression = "0 2 * * 0" // 每周日2点执行一次
	}
	if backupScheduleSetting.Enable {
		if err := t.ScheduleBackupTask(backupScheduleSetting.CronExpression); err != nil {
			return err
		}
	}
	if err := t.restoreVisitorStats(context.Background()); err != nil {
		return err
	}
	if err := t.VisitorSnapshotTask(); err != nil {
		return err
	}

	t.scheduler.Start()
	t.started = true
	return nil
}

func (t *Tasker) Stop(context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.started || t.scheduler == nil {
		return nil
	}
	if err := t.scheduler.Shutdown(); err != nil {
		logUtil.GetLogger().Error("Failed to shutdown scheduler", zap.Error(err))
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
				if err := t.fileService.CleanupOrphanFiles(); err != nil {
					logUtil.GetLogger().
						Error("Failed to clean up temporary files", zap.Error(err))
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule CleanupTempFilesTask", zap.Error(err))
		return err
	}
	return nil
}

// DeadLetterConsumeTask 死信任务消费任务
func (t *Tasker) DeadLetterConsumeTask() error {
	// 每5分钟执行一次，保证死信重试能在分钟级恢复
	_, err := t.scheduler.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(
			func() {
				// 取出死信队列中的任务，逐个重试
				deadLetters, err := t.queueRepo.ListDeadLetters(context.Background(), 10)
				if err != nil {
					logUtil.GetLogger().
						Error("Failed To Get DeadLetters!", zap.Error(err))
				}

				// 遍历死信任务，重新发送事件
				for _, dl := range deadLetters {
					// 发布事件到事件总线，触发重试
					if err := t.publisher.DeadLetterRetried(
						context.Background(),
						contracts.DeadLetterRetriedEvent{DeadLetter: dl},
					); err != nil {
						logUtil.GetLogger().
							Error("Failed to publish dead letter retried event", zap.Error(err))
					}
				}
			},
		),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule WebhookRetryTask", zap.Error(err))
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
				ctx := context.Background()

				backupPath, fileName, err := backup.ExecuteBackup()
				if err != nil {
					logUtil.GetLogger().Error("Failed to execute scheduled backup",
						zap.String("path", backupPath),
						zap.String("fileName", fileName),
						zap.Error(err))
					return
				}

				t.tryUploadBackupToS3(ctx, backupPath, fileName)

				if err := t.publisher.SystemBackup(
					ctx,
					contracts.SystemBackupEvent{Info: "System scheduled backup completed"},
				); err != nil {
					logUtil.GetLogger().
						Error("Failed to publish backup completed event", zap.Error(err))
				}
			},
		),
		gocron.WithTags(backupScheduleTag),
	)
	if err != nil {
		logUtil.GetLogger().
			Error("Failed to schedule ScheduleBackupTask", zap.Error(err))
		return err
	}
	return nil
}

// VisitorSnapshotTask 定时保存 PV/UV 快照并清理历史数据。
func (t *Tasker) VisitorSnapshotTask() error {
	_, err := t.scheduler.NewJob(
		gocron.CronJob("59 23 * * *", false),
		gocron.NewTask(func() {
			if t.visitorTracker == nil || t.visitorRepo == nil {
				return
			}
			todayStat := t.visitorTracker.TodayStat()
			if err := t.visitorRepo.UpsertDailyStat(context.Background(), buildVisitorDailyStat(todayStat)); err != nil {
				logUtil.GetLogger().Error("Failed to upsert visitor daily stat", zap.Error(err))
				return
			}
			// 保留最近 6 天（包括当天快照）。
			cutoff := visitorCutoffDate(time.Now().UTC())
			if err := t.visitorRepo.DeleteOlderThan(context.Background(), cutoff); err != nil {
				logUtil.GetLogger().Error("Failed to cleanup visitor stats", zap.Error(err))
			}
		}),
		gocron.WithTags(visitorSnapshotTag),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule VisitorSnapshotTask", zap.Error(err))
		return err
	}
	return nil
}

func (t *Tasker) restoreVisitorStats(ctx context.Context) error {
	if t.visitorTracker == nil || t.visitorRepo == nil {
		return nil
	}
	stats, err := t.visitorRepo.GetRecentDays(ctx, 6)
	if err != nil {
		logUtil.GetLogger().Error("Failed to load visitor stats", zap.Error(err))
		return err
	}
	if len(stats) == 0 {
		return nil
	}
	t.visitorTracker.LoadHistory(convertVisitorHistory(stats))
	return nil
}

func buildVisitorDailyStat(stat visitor.DayStat) visitorModel.DailyStat {
	return visitorModel.DailyStat{
		Date: stat.Date,
		PV:   stat.PV,
		UV:   stat.UV,
	}
}

func convertVisitorHistory(stats []visitorModel.DailyStat) []visitor.DayStat {
	history := make([]visitor.DayStat, 0, len(stats))
	for _, stat := range stats {
		history = append(history, visitor.DayStat{
			Date: stat.Date,
			PV:   stat.PV,
			UV:   stat.UV,
		})
	}
	return history
}

func visitorCutoffDate(now time.Time) string {
	return now.UTC().AddDate(0, 0, -5).Format("2006-01-02")
}

// ApplyBackupSchedule 在运行期动态更新备份任务。
func (t *Tasker) ApplyBackupSchedule(schedule settingModel.BackupSchedule) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.scheduler == nil {
		return errors.New("scheduler is nil")
	}
	if !t.started {
		return nil
	}

	logUtil.GetLogger().Info("Applying backup schedule",
		zap.Bool("enable", schedule.Enable),
		zap.String("cron", schedule.CronExpression),
	)

	// 先移除旧任务，避免重复触发。
	t.scheduler.RemoveByTags(backupScheduleTag)
	if !schedule.Enable {
		logUtil.GetLogger().Info("Backup schedule disabled, jobs removed")
		return nil
	}

	if err := t.ScheduleBackupTask(schedule.CronExpression); err != nil {
		logUtil.GetLogger().Error("Failed to apply backup schedule", zap.Error(err))
		return err
	}
	logUtil.GetLogger().Info("Backup schedule applied successfully")
	return nil
}

func (t *Tasker) tryUploadBackupToS3(ctx context.Context, backupPath, fileName string) {
	if t.storageManager == nil {
		return
	}
	selector := t.storageManager.GetSelector()
	if selector == nil || !selector.ObjectEnabled() {
		return
	}

	cfg := t.storageManager.GetStorageConfig(ctx)
	if err := backup.UploadToS3(ctx, backupPath, fileName, cfg); err != nil {
		logUtil.GetLogger().Warn("Failed to upload backup to S3", zap.Error(err))
	}
}
