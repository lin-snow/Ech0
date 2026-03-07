package model

import (
	"time"
)

const (
	// DeadLetterTypeWebhook webhook 类型的死信任务
	DeadLetterTypeWebhook = "webhook"
)

const (
	// DeadLetterStatusPending 待处理状态
	DeadLetterStatusPending = "pending"
	// DeadLetterStatusProcessing 处理中状态
	DeadLetterStatusProcessing = "processing"
	// DeadLetterStatusFailed 失败状态
	DeadLetterStatusFailed = "failed"
	// DeadLetterStatusCompleted 完成状态
	DeadLetterStatusCompleted = "completed"
	// DeadLetterStatusDiscarded 丢弃状态
	DeadLetterStatusDiscarded = "discarded"
)

const (
	DeadLetterMetaKey = "dead_letter" // 存储在 Payload 中的元数据键
)

// ReplayPayload 重放任务的载荷，使用 map 以支持灵活的字段
type ReplayPayload map[string]any

// DeadLetter 死信任务模型
type DeadLetter struct {
	ID         int64     `gorm:"primaryKey"  json:"id"`          // 任务 ID
	Type       string    `gorm:"type"        json:"type"`        // 业务类型，如 "webhook" / "push" / "email"
	Payload    []byte    `gorm:"payload"     json:"payload"`     // 原始任务数据（序列化 JSON）
	ErrorMsg   string    `gorm:"error_msg"   json:"error_msg"`   // 失败原因（错误信息）
	RetryCount int       `gorm:"retry_count" json:"retry_count"` // 重试次数
	NextRetry  time.Time `gorm:"next_retry"  json:"next_retry"`  // 下次重试时间（指数退避）
	Status     string    `gorm:"status"      json:"status"`      // 任务状态，如 "pending", "processing", "failed", "completed","discarded"
	CreatedAt  time.Time `gorm:"created_at"  json:"created_at"`
	UpdatedAt  time.Time `gorm:"updated_at"  json:"updated_at"`
}

func (dl *DeadLetter) SetType(t string) {
	dl.Type = t
}
