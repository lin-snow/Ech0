// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package outbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	outboxModel "github.com/lin-snow/ech0/internal/model/outbox"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	"github.com/lin-snow/ech0/internal/transaction"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/internal/webhook"
	"gorm.io/gorm"
)

type Worker struct {
	mailer   commentService.Mailer
	settings kvstore.Store
	db       func() *gorm.DB
	index    *embeddingService.EmbeddingService
	send     func(context.Context, *webhookModel.Webhook, event.WebhookObservation, string) error
}

func New(db func() *gorm.DB, index *embeddingService.EmbeddingService, sender *webhook.Sender, mailer *commentService.GoMailSender, settings kvstore.Store) *Worker {
	return &Worker{db: db, index: index, send: sender.DeliverOnce, mailer: mailer, settings: settings}
}

func (w *Worker) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		eventbus.On(w.queueMail),
		record[event.EchoCreated](w), record[event.EchoUpdated](w), record[event.EchoDeleted](w),
		record[event.UserCreated](w), record[event.UserUpdated](w), record[event.UserDeleted](w),
		record[event.CommentCreated](w), record[event.CommentStatusUpdated](w), record[event.CommentDeleted](w),
		record[event.ResourceUploaded](w), record[event.SystemSnapshot](w), record[event.SystemExport](w), record[event.UpdateSnapshotSchedule](w),
	}
}

func record[T event.Named](w *Worker) eventbus.Registration {
	return eventbus.OnWithMeta(func(ctx context.Context, e T, meta map[string]string) error {
		if tx, ok := transaction.TxFromContext(ctx); ok {
			return w.record(ctx, tx, e, meta)
		}
		return w.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error { return w.record(ctx, tx, e, meta) })
	})
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (w *Worker) record(ctx context.Context, tx *gorm.DB, e event.Named, meta map[string]string) error {
	tx = tx.WithContext(ctx)
	rows := make([]outboxModel.Delivery, 0)
	echoChange := false
	switch e.(type) {
	case event.EchoCreated, event.EchoUpdated, event.EchoDeleted:
		echoChange = true
	}
	if echoChange {
		rows = append(rows, outboxModel.Delivery{ID: uuidUtil.NewV7(), Lane: "embedding", Subject: e.(event.Keyed).OrderingKey()})
	}
	if echoChange || e.EventName() == "user.deleted" {
		if err := tx.Exec("DELETE FROM key_values WHERE key=?", string(agent.GEN_RECENT)).Error; err != nil {
			return err
		}
	}
	var hooks []webhookModel.Webhook
	if err := tx.Select("id,url").Where("is_active=?", true).Find(&hooks).Error; err != nil {
		return err
	}
	if len(hooks) > 0 {
		observation, err := event.NewWebhookObservation(e.EventName(), e, meta)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(observation)
		if err != nil {
			return err
		}
		for _, hook := range hooks {
			rows = append(rows, outboxModel.Delivery{ID: uuidUtil.NewV7(), Lane: "webhook/" + hook.ID, Subject: hook.ID, Destination: digest(hook.URL), Payload: payload})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.CreateInBatches(rows, 50).Error
}

type Ticket struct {
	ID      string `json:"effect"`
	Attempt int64  `json:"attempt"`
	Due     int64  `json:"effect_due"`
}

func (t Ticket) Key() string {
	return "effect-" + t.ID + "-" + strconv.FormatInt(t.Attempt, 10) + "-" + strconv.FormatInt(t.Due, 10)
}

const headOfLane = "NOT EXISTS (SELECT 1 FROM ech0_deliveries older WHERE older.lane=ech0_deliveries.lane AND older.rowid<ech0_deliveries.rowid)"

func (w *Worker) Next(ctx context.Context) (*Ticket, error) {
	var t Ticket
	res := w.db().WithContext(ctx).Raw("SELECT id,attempt,ready AS due FROM ech0_deliveries WHERE " + headOfLane + " ORDER BY ready,rowid LIMIT 1").Scan(&t)
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, res.Error
	}
	return &t, nil
}

// Each invocation holds a fenced, expiring claim for one delivery. An HTTP
// timeout is ambiguous, so every attempt uses the same webhook event ID.
func (w *Worker) Advance(ctx context.Context, t Ticket) error {
	db := w.db().WithContext(ctx)
	lease := uuidUtil.NewV7()
	var d outboxModel.Delivery
	res := db.Raw("UPDATE ech0_deliveries SET lease=?,ready=unixepoch()+90 WHERE id=? AND attempt=? AND ready<=unixepoch() AND "+headOfLane+" RETURNING *", lease, t.ID, t.Attempt).Scan(&d)
	if res.Error != nil || res.RowsAffected == 0 {
		return res.Error
	}
	workCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	if d.Lane == "embedding" && d.Subject == reindexSubject {
		return w.reindex(workCtx, db, d)
	}
	deliveryErr := w.deliver(workCtx, db.WithContext(workCtx), d)
	return db.Transaction(func(tx *gorm.DB) error {
		claim := tx.Where("id=? AND lease=?", d.ID, lease)
		retry := deliveryErr != nil && !errors.Is(deliveryErr, errRetired)
		if retry && (d.Attempt < 3 || !strings.HasPrefix(d.Lane, "webhook/")) {
			delay := time.Duration(1<<min(d.Attempt, 12)) * time.Second
			return claim.Model(&outboxModel.Delivery{}).Updates(map[string]any{"attempt": d.Attempt + 1, "ready": time.Now().Add(min(delay, time.Hour)).Unix(), "lease": ""}).Error
		}
		done := claim.Delete(&outboxModel.Delivery{})
		if done.Error != nil || done.RowsAffected == 0 || errors.Is(deliveryErr, errRetired) || !strings.HasPrefix(d.Lane, "webhook/") {
			return done.Error
		}
		status := "success"
		if deliveryErr != nil {
			status = "failed"
		}
		return tx.Model(&webhookModel.Webhook{}).Where("id=? AND is_active=?", d.Subject, true).Updates(map[string]any{"last_status": status, "last_trigger": time.Now().Unix()}).Error
	})
}

var errRetired = errors.New("webhook was retired")

func (w *Worker) deliver(ctx context.Context, db *gorm.DB, d outboxModel.Delivery) error {
	if strings.HasPrefix(d.Lane, "mail/") {
		var msg commentService.MailMessage
		if err := json.Unmarshal(d.Payload, &msg); err != nil {
			return err
		}
		msg.ID = d.ID
		return commentService.DeliverQueuedMail(ctx, w.settings, w.mailer, msg)
	}
	if d.Lane == "embedding" {
		var echo echoModel.Echo
		err := db.Preload("Tags").First(&echo, "id=?", d.Subject).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return w.index.RemoveEcho(ctx, d.Subject)
		}
		if err != nil {
			return err
		}
		return w.index.IndexEcho(ctx, echo)
	}
	var hook webhookModel.Webhook
	err := db.First(&hook, "id=? AND is_active=?", d.Subject, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && digest(hook.URL) != d.Destination) {
		return errRetired
	}
	if err != nil {
		return err
	}
	var observation event.WebhookObservation
	if err := json.Unmarshal(d.Payload, &observation); err != nil {
		return err
	}
	return w.send(ctx, &hook, observation, d.ID)
}

func (w *Worker) SnapshotCompleted(ctx context.Context, tx *gorm.DB, scheduled bool) error {
	info := "System manual snapshot completed"
	if scheduled {
		info = "System scheduled snapshot completed"
	}
	return w.record(ctx, tx, event.SystemSnapshot{Info: info}, map[string]string{"source": "ech0"})
}

func (w *Worker) queueMail(ctx context.Context, e event.CommentMail) error {
	tx, ok := transaction.TxFromContext(ctx)
	if !ok {
		return errors.New("mail delivery requires a mutation transaction")
	}
	payload, err := json.Marshal(e.Message)
	if err != nil {
		return err
	}
	return tx.WithContext(ctx).Create(&outboxModel.Delivery{ID: uuidUtil.NewV7(), Lane: "mail/" + digest(e.Message.To), Payload: payload}).Error
}
