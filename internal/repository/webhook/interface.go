package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/webhook"
)

type WebhookRepositoryInterface interface {
	// CreateWebhook 创建一个webhook
	CreateWebhook(ctx context.Context, webhook *model.Webhook) error

	// GetAllWebhooks 获取所有webhooks
	GetAllWebhooks(ctx context.Context) ([]model.Webhook, error)

	// DeleteWebhookByID 根据ID删除webhook
	DeleteWebhookByID(ctx context.Context, id uint) error

	// ListActiveWebhooks 列出所有激活的 webhook
	ListActiveWebhooks(ctx context.Context) ([]model.Webhook, error)
}
