// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	captchaCfg "github.com/lin-snow/ech0/internal/captcha"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const (
	minSubmitMS        int64 = 2000
	maxFormTokenHours  int64 = 24
	maxCommentRunes          = 200
	recentDuplicateSec int64 = 90

	integrationShortWindow int64 = 60
	integrationLongWindow  int64 = 3600
	integrationShortLimit  int64 = 5
	integrationLongLimit   int64 = 30
)

type CommentService struct {
	commonService CommonService
	repo          Repository
	durableKV     kvstore.Store
	bus           *busen.Bus
	mailer        Mailer
}

func NewCommentService(
	commonService CommonService,
	repo Repository,
	durableKV kvstore.Store,
	busProvider func() *busen.Bus,
	mailer Mailer,
) *CommentService {
	return &CommentService{
		commonService: commonService,
		repo:          repo,
		durableKV:     durableKV,
		bus:           busProvider(),
		mailer:        mailer,
	}
}

func (s *CommentService) GetFormMeta(ctx context.Context, clientIP, apiBaseURL string) (model.FormMeta, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return model.FormMeta{}, err
	}
	captchaAPIEndpoint := captchaCfg.APIEndpointWithBase(apiBaseURL)
	captchaReady := setting.CaptchaEnabled &&
		captchaAPIEndpoint != "" &&
		strings.TrimSpace(captchaCfg.Secret()) != ""
	issuedAt := time.Now().UnixMilli()
	token := s.signFormToken(clientIP, issuedAt)
	return model.FormMeta{
		FormToken:          token,
		MinSubmitMs:        minSubmitMS,
		CaptchaEnabled:     captchaReady,
		CaptchaAPIEndpoint: captchaAPIEndpoint,
		EnableComment:      setting.EnableComment,
	}, nil
}

func (s *CommentService) CreateComment(
	ctx context.Context,
	clientIP string,
	userAgent string,
	dto *model.CreateCommentDto,
) (model.CreateCommentResult, error) {
	if strings.TrimSpace(dto.HoneypotField) != "" {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "提交被拒绝")
	}
	if err := s.verifyFormToken(clientIP, dto.FormToken); err != nil {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "提交过快或表单已失效")
	}

	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return model.CreateCommentResult{}, err
	}
	if !setting.EnableComment {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论功能未启用")
	}

	if setting.CaptchaEnabled {
		if err := captchaCfg.SiteVerify(dto.CaptchaToken); err != nil {
			return model.CreateCommentResult{},
				commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "验证码验证失败")
		}
	}

	user, validUser, err := s.resolveRequestUser(ctx)
	if err != nil {
		return model.CreateCommentResult{}, err
	}

	comment := model.Comment{
		EchoID:    strings.TrimSpace(dto.EchoID),
		Content:   strings.TrimSpace(dto.Content),
		IPHash:    hashClientIP(clientIP),
		UserAgent: strings.TrimSpace(userAgent),
		Status:    model.StatusPending,
		Source:    model.SourceGuest,
	}
	if comment.EchoID == "" || comment.Content == "" {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能为空")
	}
	if utf8.RuneCountInString(comment.Content) > maxCommentRunes {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能超过200字")
	}

	parentID, err := s.resolveParentID(ctx, comment.EchoID, dto.ParentID)
	if err != nil {
		return model.CreateCommentResult{}, err
	}
	comment.ParentID = parentID

	if validUser && (user.IsAdmin || user.IsOwner) {
		comment.Source = model.SourceSystem
		comment.Nickname = user.Username
		comment.Email = ""
		comment.UserID = &user.ID
		comment.Status = model.StatusApproved
	} else {
		nickname := strings.TrimSpace(dto.Nickname)
		email := strings.TrimSpace(dto.Email)
		website := strings.TrimSpace(dto.Website)
		if nickname == "" || email == "" {
			return model.CreateCommentResult{},
				commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "昵称和邮箱不能为空")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return model.CreateCommentResult{},
				commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "邮箱格式无效")
		}
		if website != "" {
			parsed, err := url.ParseRequestURI(website)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return model.CreateCommentResult{},
					commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "网址格式无效")
			}
		}
		comment.Nickname = nickname
		comment.Email = email
		comment.Website = website

		if !setting.RequireApproval {
			comment.Status = model.StatusApproved
		}

		if err := s.checkRateLimit(ctx, comment.IPHash, email, ""); err != nil {
			return model.CreateCommentResult{}, err
		}
	}

	duplicated, err := s.repo.ExistsRecentDuplicate(
		ctx,
		comment.EchoID,
		comment.Content,
		comment.Email,
		comment.IPHash,
		derefString(comment.UserID),
		recentDuplicateSec,
	)
	if err != nil {
		return model.CreateCommentResult{}, err
	}
	if duplicated {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "请勿重复提交相同评论")
	}

	if err := s.repo.CreateComment(ctx, &comment); err != nil {
		return model.CreateCommentResult{}, err
	}
	s.emitCommentCreated(ctx, comment)
	if shouldNotifyOwnerOnCreate(comment.Source) {
		s.notifyOwnerAsync(ctx, "created", comment)
	}
	s.notifyReplyTargetAsync(ctx, comment)
	return model.CreateCommentResult{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (s *CommentService) resolveParentID(ctx context.Context, echoID, rawParentID string) (*string, error) {
	parentID := strings.TrimSpace(rawParentID)
	if parentID == "" {
		return nil, nil
	}
	parent, err := s.repo.GetCommentByID(ctx, parentID)
	if err != nil || parent.ID == "" || parent.EchoID != echoID {
		return nil, commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "回复的评论不存在")
	}
	if parent.Status != model.StatusApproved {
		return nil, commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "该评论暂不可回复")
	}
	return &parent.ID, nil
}

func (s *CommentService) CreateIntegrationComment(
	ctx context.Context,
	clientIP string,
	userAgent string,
	dto *model.CreateIntegrationCommentDto,
) (model.CreateCommentResult, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return model.CreateCommentResult{}, err
	}
	if !setting.EnableComment {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论功能未启用")
	}

	v := viewer.MustFromContext(ctx)
	userID := strings.TrimSpace(v.UserID())
	tokenID := strings.TrimSpace(v.TokenID())

	comment := model.Comment{
		EchoID:    strings.TrimSpace(dto.EchoID),
		Content:   strings.TrimSpace(dto.Content),
		IPHash:    hashClientIP(clientIP),
		UserAgent: strings.TrimSpace(userAgent),
		Status:    model.StatusPending,
		Source:    model.SourceIntegration,
	}
	if comment.EchoID == "" || comment.Content == "" {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能为空")
	}
	if utf8.RuneCountInString(comment.Content) > maxCommentRunes {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论内容不能超过200字")
	}

	nickname := strings.TrimSpace(dto.Nickname)
	if nickname == "" {
		nickname = "Integration"
	}
	comment.Nickname = nickname

	if userID != "" {
		comment.UserID = &userID
	}

	if !setting.RequireApproval {
		comment.Status = model.StatusApproved
	}

	if err := s.checkIntegrationRateLimit(ctx, comment.IPHash, userID, tokenID); err != nil {
		return model.CreateCommentResult{}, err
	}

	duplicated, err := s.repo.ExistsRecentDuplicate(
		ctx,
		comment.EchoID,
		comment.Content,
		comment.Email,
		comment.IPHash,
		derefString(comment.UserID),
		recentDuplicateSec,
	)
	if err != nil {
		return model.CreateCommentResult{}, err
	}
	if duplicated {
		return model.CreateCommentResult{},
			commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "请勿重复提交相同评论")
	}

	if err := s.repo.CreateComment(ctx, &comment); err != nil {
		return model.CreateCommentResult{}, err
	}

	logUtil.GetLogger().Info("integration comment created",
		slog.String("comment_id", comment.ID),
		slog.String("echo_id", comment.EchoID),
		slog.String("token_id", tokenID),
		slog.String("user_id", userID),
		slog.String("source", string(comment.Source)),
		slog.String("metadata", strings.TrimSpace(dto.Metadata)),
	)

	s.emitCommentCreated(ctx, comment)
	if shouldNotifyOwnerOnCreate(comment.Source) {
		s.notifyOwnerAsync(ctx, "created", comment)
	}
	return model.CreateCommentResult{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (s *CommentService) checkIntegrationRateLimit(ctx context.Context, ipHash, userID, _ string) error {
	ipShort, err := s.repo.CountByIPWithin(ctx, ipHash, integrationShortWindow)
	if err != nil {
		return err
	}
	ipLong, err := s.repo.CountByIPWithin(ctx, ipHash, integrationLongWindow)
	if err != nil {
		return err
	}
	if ipShort >= integrationShortLimit || ipLong >= integrationLongLimit {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "集成评论过于频繁，请稍后再试")
	}

	if userID != "" {
		userShort, err := s.repo.CountByUserWithin(ctx, userID, integrationShortWindow)
		if err != nil {
			return err
		}
		if userShort >= integrationShortLimit {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "集成评论过于频繁，请稍后再试")
		}
	}

	return nil
}

func (s *CommentService) ListPublicByEchoID(ctx context.Context, echoID string) ([]model.PublicComment, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.EnableComment {
		return []model.PublicComment{}, nil
	}
	rows, err := s.repo.ListPublicByEchoID(ctx, strings.TrimSpace(echoID))
	if err != nil {
		return nil, err
	}
	return model.ToPublicComments(rows), nil
}

func (s *CommentService) ListPublicComments(ctx context.Context, limit int) ([]model.PublicComment, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.EnableComment {
		return []model.PublicComment{}, nil
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.repo.ListPublicComments(ctx, limit)
	if err != nil {
		return nil, err
	}
	return model.ToPublicComments(rows), nil
}

func (s *CommentService) ListPanelComments(
	ctx context.Context,
	query model.ListCommentQuery,
) (model.PageResult[model.Comment], error) {
	if err := s.requireAdmin(ctx); err != nil {
		return model.PageResult[model.Comment]{}, err
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.repo.ListComments(ctx, query)
}

func (s *CommentService) GetCommentByID(ctx context.Context, id string) (model.Comment, error) {
	if err := s.requireAdmin(ctx); err != nil {
		return model.Comment{}, err
	}
	return s.repo.GetCommentByID(ctx, id)
}

func (s *CommentService) UpdateCommentStatus(ctx context.Context, id string, status model.Status) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if status != model.StatusPending && status != model.StatusApproved && status != model.StatusRejected {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "无效的评论状态")
	}
	if err := s.repo.UpdateCommentStatus(ctx, id, status); err != nil {
		return err
	}
	if updated, err := s.repo.GetCommentByID(ctx, id); err == nil && updated.ID != "" {
		s.emitCommentStatusUpdated(ctx, updated)
		s.notifyOwnerAsync(ctx, "status", updated)
	}
	return nil
}

func (s *CommentService) UpdateCommentHot(ctx context.Context, id string, hot bool) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if err := s.repo.UpdateCommentHot(ctx, id, hot); err != nil {
		return err
	}
	if hot {
		if updated, err := s.repo.GetCommentByID(ctx, id); err == nil && updated.ID != "" {
			s.notifyOwnerAsync(ctx, "hot", updated)
		}
	}
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, id string) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	beforeDelete, _ := s.repo.GetCommentByID(ctx, id)
	if err := s.repo.DeleteComment(ctx, id); err != nil {
		return err
	}
	if beforeDelete.ID != "" {
		s.emitCommentDeleted(ctx, beforeDelete)
	}
	return nil
}

func (s *CommentService) BatchAction(ctx context.Context, action string, ids []string) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	switch action {
	case "approve":
		if err := s.repo.BatchUpdateStatus(ctx, ids, model.StatusApproved); err != nil {
			return err
		}
		for _, id := range ids {
			if updated, err := s.repo.GetCommentByID(ctx, id); err == nil && updated.ID != "" {
				s.emitCommentStatusUpdated(ctx, updated)
				s.notifyOwnerAsync(ctx, "status", updated)
			}
		}
		return nil
	case "reject":
		if err := s.repo.BatchUpdateStatus(ctx, ids, model.StatusRejected); err != nil {
			return err
		}
		for _, id := range ids {
			if updated, err := s.repo.GetCommentByID(ctx, id); err == nil && updated.ID != "" {
				s.emitCommentStatusUpdated(ctx, updated)
				s.notifyOwnerAsync(ctx, "status", updated)
			}
		}
		return nil
	case "delete":
		beforeDelete := make([]model.Comment, 0, len(ids))
		for _, id := range ids {
			if comment, err := s.repo.GetCommentByID(ctx, id); err == nil && comment.ID != "" {
				beforeDelete = append(beforeDelete, comment)
			}
		}
		if err := s.repo.BatchDelete(ctx, ids); err != nil {
			return err
		}
		for _, comment := range beforeDelete {
			s.emitCommentDeleted(ctx, comment)
		}
		return nil
	default:
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "无效的批量动作")
	}
}

func (s *CommentService) emitCommentCreated(ctx context.Context, comment model.Comment) {
	if comment.ID != "" {
		eventbus.Notify(ctx, s.bus, event.CommentCreated{Comment: comment})
	}
}

func (s *CommentService) emitCommentStatusUpdated(ctx context.Context, comment model.Comment) {
	if comment.ID != "" {
		eventbus.Notify(ctx, s.bus, event.CommentStatusUpdated{Comment: comment})
	}
}

func (s *CommentService) emitCommentDeleted(ctx context.Context, comment model.Comment) {
	if comment.ID != "" {
		eventbus.Notify(ctx, s.bus, event.CommentDeleted{Comment: comment})
	}
}

func (s *CommentService) GetSystemSetting(ctx context.Context) (model.SystemSetting, error) {
	setting, err := s.getSystemSettingRaw(ctx)
	if err != nil {
		return model.SystemSetting{}, err
	}
	return sanitizeSettingForOutput(setting), nil
}

func (s *CommentService) getSystemSettingRaw(ctx context.Context) (model.SystemSetting, error) {
	return coreSetting.Get(ctx, s.durableKV, coreSetting.Comment)
}

func (s *CommentService) UpdateSystemSetting(ctx context.Context, setting model.SystemSetting) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	applySettingDefaults(&setting)
	current, err := s.getSystemSettingRaw(ctx)
	if err == nil && strings.TrimSpace(setting.EmailNotify.SMTPPassword) == "" {
		setting.EmailNotify.SMTPPassword = current.EmailNotify.SMTPPassword
	}
	buf, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	return s.durableKV.Set(ctx, model.CommentSystemSettingKey, string(buf))
}

func (s *CommentService) SendTestEmail(ctx context.Context, setting model.SystemSetting) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	applySettingDefaults(&setting)
	current, err := s.getSystemSettingRaw(ctx)
	if err == nil && strings.TrimSpace(setting.EmailNotify.SMTPPassword) == "" {
		setting.EmailNotify.SMTPPassword = current.EmailNotify.SMTPPassword
	}
	ownerEmail, err := s.resolveOwnerEmail()
	if err != nil {
		return err
	}
	serverURL := s.resolveServerURL(ctx)
	if err := validateEmailNotifySetting(setting.EmailNotify, ownerEmail); err != nil {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, err.Error())
	}
	content := buildNotifyContent("test", model.Comment{
		EchoID:    "demo",
		Nickname:  "comment-test",
		Email:     "author@example.com",
		Content:   "这是一条测试评论内容，用于预览 Ech0 邮件通知样式。",
		Status:    model.StatusPending,
		CreatedAt: time.Now().UTC().Unix(),
	}, serverURL)
	return s.sendOwnerMail(ctx, setting.EmailNotify, MailMessage{
		To:       ownerEmail,
		Subject:  content.Subject,
		TextBody: content.TextBody,
		HTMLBody: content.HTMLBody,
	})
}

func (s *CommentService) notifyOwnerAsync(ctx context.Context, kind string, comment model.Comment) {
	setting, err := s.getSystemSettingRaw(ctx)
	if err != nil {
		return
	}
	if !shouldNotify(setting, kind, comment.Status) {
		return
	}
	recipient := ""
	if useCommentRecipient(kind) {
		var ok bool
		recipient, ok = parseValidEmail(comment.Email)
		if !ok {
			logUtil.GetLogger().Info("skip comment notify mail due to invalid commenter email",
				slog.String("kind", kind),
				slog.String("comment_id", comment.ID))
			return
		}
	} else {
		ownerEmail, ownerErr := s.resolveOwnerEmail()
		if ownerErr != nil {
			return
		}
		recipient = ownerEmail
	}
	serverURL := s.resolveServerURL(ctx)
	content := buildNotifyContent(kind, comment, serverURL)
	go func(cfg model.EmailNotifySetting, msg MailMessage) {
		notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.sendOwnerMail(notifyCtx, cfg, msg); err != nil {
			logUtil.GetLogger().Warn("comment notify mail failed", logUtil.Err(err), slog.String("comment_id", comment.ID))
		}
	}(setting.EmailNotify, MailMessage{
		To:       recipient,
		Subject:  content.Subject,
		TextBody: content.TextBody,
		HTMLBody: content.HTMLBody,
	})
}

func (s *CommentService) notifyReplyTargetAsync(ctx context.Context, comment model.Comment) {
	if comment.ParentID == nil || strings.TrimSpace(*comment.ParentID) == "" {
		return
	}
	setting, err := s.getSystemSettingRaw(ctx)
	if err != nil || !setting.EmailNotify.Enabled {
		return
	}
	parent, err := s.repo.GetCommentByID(ctx, strings.TrimSpace(*comment.ParentID))
	if err != nil || parent.ID == "" {
		return
	}
	targetEmail, ok := parseValidEmail(parent.Email)
	if !ok {
		return
	}
	if strings.EqualFold(targetEmail, strings.TrimSpace(comment.Email)) {
		return
	}
	if ownerEmail, ownerErr := s.resolveOwnerEmail(); ownerErr == nil &&
		strings.EqualFold(targetEmail, strings.TrimSpace(ownerEmail)) {
		return
	}
	serverURL := s.resolveServerURL(ctx)
	content := buildNotifyContent("reply", comment, serverURL)
	go func(cfg model.EmailNotifySetting, msg MailMessage) {
		notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.sendOwnerMail(notifyCtx, cfg, msg); err != nil {
			logUtil.GetLogger().Warn("comment reply notify mail failed",
				logUtil.Err(err), slog.String("comment_id", comment.ID))
		}
	}(setting.EmailNotify, MailMessage{
		To:       targetEmail,
		Subject:  content.Subject,
		TextBody: content.TextBody,
		HTMLBody: content.HTMLBody,
	})
}

func useCommentRecipient(kind string) bool {
	return kind == "status" || kind == "hot"
}

func shouldNotifyOwnerOnCreate(source model.SourceType) bool {
	return source != model.SourceSystem
}

func parseValidEmail(raw string) (string, bool) {
	email := strings.TrimSpace(raw)
	if email == "" {
		return "", false
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", false
	}
	return email, true
}

func (s *CommentService) sendOwnerMail(ctx context.Context, cfg model.EmailNotifySetting, msg MailMessage) error {
	if s.mailer == nil {
		return errors.New("mailer unavailable")
	}
	if err := validateEmailNotifySetting(cfg, msg.To); err != nil {
		return err
	}
	return s.mailer.Send(ctx, MailerConfig{
		Host:     strings.TrimSpace(cfg.SMTPHost),
		Port:     cfg.SMTPPort,
		Username: strings.TrimSpace(cfg.SMTPUsername),
		Password: cfg.SMTPPassword,
		Sender:   strings.TrimSpace(cfg.SMTPSender),
	}, msg)
}

func shouldNotify(setting model.SystemSetting, kind string, status model.Status) bool {
	if !setting.EmailNotify.Enabled {
		return false
	}
	switch kind {
	case "created":
		return true
	case "hot":
		return true
	case "status":
		if status == model.StatusRejected {
			return true
		}
		if status == model.StatusApproved {
			return setting.RequireApproval
		}
	}
	return false
}

func validateEmailNotifySetting(cfg model.EmailNotifySetting, ownerEmail string) error {
	if strings.TrimSpace(ownerEmail) == "" {
		return errors.New("owner 邮箱不能为空")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(ownerEmail)); err != nil {
		return errors.New("owner 邮箱格式无效")
	}
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return errors.New("SMTP Host 不能为空")
	}
	if cfg.SMTPPort <= 0 {
		return errors.New("SMTP Port 无效")
	}
	if strings.TrimSpace(cfg.SMTPUsername) == "" {
		return errors.New("SMTP Username 不能为空")
	}
	sender := strings.TrimSpace(cfg.SMTPSender)
	if sender == "" {
		sender = strings.TrimSpace(cfg.SMTPUsername)
	}
	if _, err := mail.ParseAddress(sender); err != nil {
		return errors.New("发件人邮箱格式无效（请检查 SMTP 发件人地址）")
	}
	if strings.TrimSpace(cfg.SMTPPassword) == "" {
		return errors.New("SMTP Password 不能为空")
	}
	return nil
}

func applySettingDefaults(setting *model.SystemSetting) {
	if setting == nil {
		return
	}
	if setting.EmailNotify.SMTPPort <= 0 {
		setting.EmailNotify.SMTPPort = 587
	}
}

func sanitizeSettingForOutput(in model.SystemSetting) model.SystemSetting {
	out := in
	out.EmailNotify.SMTPPasswordSet = strings.TrimSpace(out.EmailNotify.SMTPPassword) != ""
	out.EmailNotify.SMTPPassword = ""
	return out
}

func (s *CommentService) resolveOwnerEmail() (string, error) {
	owner, err := s.commonService.GetOwner()
	if err != nil {
		return "", err
	}
	email := strings.TrimSpace(owner.Email)
	if email == "" {
		return "", errors.New("owner 邮箱未设置")
	}
	return email, nil
}

func (s *CommentService) resolveServerURL(ctx context.Context) string {
	if s.durableKV != nil {
		serverURL, err := s.durableKV.Get(ctx, commonModel.ServerURLKey)
		if err == nil {
			value := strings.TrimSuffix(strings.TrimSpace(serverURL), "/")
			if value != "" {
				return value
			}
		}
	}
	return strings.TrimSuffix(strings.TrimSpace(config.Config().Setting.Serverurl), "/")
}

func (s *CommentService) checkRateLimit(ctx context.Context, ipHash, email, userID string) error {
	shortWindow, longWindow := int64(60), int64(3600)
	shortLimit, longLimit := int64(3), int64(20)

	ipShort, err := s.repo.CountByIPWithin(ctx, ipHash, shortWindow)
	if err != nil {
		return err
	}
	ipLong, err := s.repo.CountByIPWithin(ctx, ipHash, longWindow)
	if err != nil {
		return err
	}
	if ipShort >= shortLimit || ipLong >= longLimit {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
	}

	emailShort, err := s.repo.CountByEmailWithin(ctx, email, shortWindow)
	if err != nil {
		return err
	}
	if emailShort >= shortLimit {
		return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
	}

	if userID != "" {
		userShort, err := s.repo.CountByUserWithin(ctx, userID, shortWindow)
		if err != nil {
			return err
		}
		if userShort >= shortLimit {
			return commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
		}
	}

	return nil
}

func (s *CommentService) verifyFormToken(clientIP, token string) error {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return errors.New("token invalid")
	}
	issuedAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}
	nowMS := time.Now().UnixMilli()
	if nowMS-issuedAt < minSubmitMS {
		return errors.New("submit too fast")
	}
	if nowMS-issuedAt > maxFormTokenHours*3600*1000 {
		return errors.New("token expired")
	}
	expectSig := s.computeHMAC(fmt.Sprintf("%s:%d", clientIP, issuedAt))
	if !hmac.Equal([]byte(parts[1]), []byte(expectSig)) {
		return errors.New("token sign mismatch")
	}
	return nil
}

func (s *CommentService) signFormToken(clientIP string, issuedAt int64) string {
	sig := s.computeHMAC(fmt.Sprintf("%s:%d", clientIP, issuedAt))
	return fmt.Sprintf("%d.%s", issuedAt, sig)
}

func (s *CommentService) computeHMAC(payload string) string {
	mac := hmac.New(sha256.New, config.Config().Security.JWTSecret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *CommentService) requireAdmin(ctx context.Context) error {
	v := viewer.MustFromContext(ctx)
	if strings.TrimSpace(v.UserID()) == "" {
		return commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	}
	user, err := s.commonService.CommonGetUserByUserId(ctx, v.UserID())
	if err != nil {
		return err
	}
	if !user.IsAdmin && !user.IsOwner {
		return commonModel.NewBizError(commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	}
	return nil
}

func (s *CommentService) resolveRequestUser(ctx context.Context) (user userModel.User, valid bool, err error) {
	v := viewer.MustFromContext(ctx)
	if strings.TrimSpace(v.UserID()) == "" {
		return user, false, nil
	}
	u, err := s.commonService.CommonGetUserByUserId(ctx, v.UserID())
	if err != nil {
		return user, false, nil
	}
	return u, true, nil
}

func ParseOptionalUserIDFromAuthHeader(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	claims, err := jwtUtil.ParseToken(strings.TrimSpace(parts[1]))
	if err != nil {
		return ""
	}
	return claims.Userid
}

func hashClientIP(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(ip)))
	return hex.EncodeToString(sum[:])
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}
