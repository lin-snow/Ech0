package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bytedance/sonic"
	"github.com/lin-snow/ech0/internal/config"
	contracts "github.com/lin-snow/ech0/internal/event/contracts"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	mailUtil "github.com/lin-snow/ech0/internal/util/mail"
	"github.com/lin-snow/ech0/pkg/viewer"
	"go.uber.org/zap"
)

const (
	minSubmitMS        int64 = 2000
	maxFormTokenHours  int64 = 24
	maxCommentRunes          = 200
	recentDuplicateSec int64 = 90
)

type CommentService struct {
	commonService      CommonService
	repo               Repository
	keyvalueRepository KeyValueRepository
	publisher          EventPublisher
}

func NewCommentService(
	commonService CommonService,
	repo Repository,
	keyvalueRepository KeyValueRepository,
	publisher EventPublisher,
) *CommentService {
	return &CommentService{
		commonService:      commonService,
		repo:               repo,
		keyvalueRepository: keyvalueRepository,
		publisher:          publisher,
	}
}

func (s *CommentService) GetFormMeta(ctx context.Context, clientIP string) (model.FormMeta, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return model.FormMeta{}, err
	}
	captchaAPIEndpoint := deriveCaptchaAPIEndpoint(setting.CaptchaVerify)
	captchaReady := setting.CaptchaEnabled &&
		captchaAPIEndpoint != "" &&
		strings.TrimSpace(setting.CaptchaSecret) != ""
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

	if setting.CaptchaEnabled && strings.TrimSpace(setting.CaptchaVerify) != "" {
		if err := s.verifyCaptcha(dto.CaptchaToken, clientIP, setting); err != nil {
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

	if validUser && (user.IsAdmin || user.IsOwner) {
		comment.Source = model.SourceSystem
		comment.Nickname = user.Username
		// 内部成员评论允许邮箱为空，不再自动填充占位邮箱。
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

	if setting.EnableEmailNotification {
		go func() {
			if err := s.sendEmailNotification(context.Background(), comment, setting); err != nil {
				logUtil.GetLogger().Error("send email notification failed", zap.Error(err))
			}
		}()
	}
	return model.CreateCommentResult{
		ID:     comment.ID,
		Status: comment.Status,
	}, nil
}

func (s *CommentService) ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.EnableComment {
		return []model.Comment{}, nil
	}
	return s.repo.ListPublicByEchoID(ctx, strings.TrimSpace(echoID))
}

func (s *CommentService) ListPublicComments(ctx context.Context, limit int) ([]model.Comment, error) {
	setting, err := s.GetSystemSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.EnableComment {
		return []model.Comment{}, nil
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListPublicComments(ctx, limit)
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
	}
	return nil
}

func (s *CommentService) UpdateCommentHot(ctx context.Context, id string, hot bool) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	return s.repo.UpdateCommentHot(ctx, id, hot)
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
	if s.publisher == nil || comment.ID == "" {
		return
	}
	_ = s.publisher.CommentCreated(ctx, contracts.CommentCreatedEvent{Comment: comment})
}

func (s *CommentService) emitCommentStatusUpdated(ctx context.Context, comment model.Comment) {
	if s.publisher == nil || comment.ID == "" {
		return
	}
	_ = s.publisher.CommentStatusUpdated(ctx, contracts.CommentStatusUpdatedEvent{Comment: comment})
}

func (s *CommentService) emitCommentDeleted(ctx context.Context, comment model.Comment) {
	if s.publisher == nil || comment.ID == "" {
		return
	}
	_ = s.publisher.CommentDeleted(ctx, contracts.CommentDeletedEvent{Comment: comment})
}

func (s *CommentService) GetSystemSetting(ctx context.Context) (model.SystemSetting, error) {
	raw, err := s.keyvalueRepository.GetKeyValue(ctx, model.CommentSystemSettingKey)
	if err != nil {
		defaultSetting := model.SystemSetting{
			EnableComment:           true,
			RequireApproval:         true,
			CaptchaEnabled:          false,
			CaptchaVerify:           "",
			CaptchaSecret:           "",
			EnableEmailNotification: false,
			SMTPHost:                "smtp.example.com",
			SMTPPort:                465,
			SMTPUser:                "",
			SMTPPassword:            "",
		}
		buf, _ := sonic.Marshal(defaultSetting)
		_ = s.keyvalueRepository.AddKeyValue(ctx, model.CommentSystemSettingKey, string(buf))
		return defaultSetting, nil
	}
	var setting model.SystemSetting
	if err := sonic.Unmarshal([]byte(raw), &setting); err != nil {
		return model.SystemSetting{}, err
	}
	return setting, nil
}

func (s *CommentService) UpdateSystemSetting(ctx context.Context, setting model.SystemSetting) error {
	if err := s.requireAdmin(ctx); err != nil {
		return err
	}
	setting.CaptchaVerify = strings.TrimSpace(setting.CaptchaVerify)
	setting.CaptchaSecret = strings.TrimSpace(setting.CaptchaSecret)
	setting.SMTPHost = strings.TrimSpace(setting.SMTPHost)
	setting.SMTPUser = strings.TrimSpace(setting.SMTPUser)
	setting.SMTPPassword = strings.TrimSpace(setting.SMTPPassword)
	buf, err := sonic.Marshal(setting)
	if err != nil {
		return err
	}
	return s.keyvalueRepository.AddOrUpdateKeyValue(ctx, model.CommentSystemSettingKey, string(buf))
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

func (s *CommentService) verifyCaptcha(token, _ string, setting model.SystemSetting) error {
	if strings.TrimSpace(token) == "" {
		return errors.New("captcha token missing")
	}
	payload := map[string]string{
		"response": strings.TrimSpace(token),
		"secret":   setting.CaptchaSecret,
	}
	body, err := sonic.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		setting.CaptchaVerify,
		strings.NewReader(string(body)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("captcha verify status: %d", resp.StatusCode)
	}

	var out struct {
		Success bool `json:"success"`
	}
	if err := sonic.ConfigFastest.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Success {
		return nil
	}
	return errors.New("captcha verify failed")
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
	if v == nil || strings.TrimSpace(v.UserID()) == "" {
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
	if v == nil || strings.TrimSpace(v.UserID()) == "" {
		return user, false, nil
	}
	u, err := s.commonService.CommonGetUserByUserId(ctx, v.UserID())
	if err != nil {
		return user, false, nil
	}
	return u, true, nil
}

func (s *CommentService) sendEmailNotification(ctx context.Context, comment model.Comment, commentSetting model.SystemSetting) error {
	client, err := mailUtil.GetSMTPClient(
		commentSetting.SMTPHost,
		commentSetting.SMTPPort,
		commentSetting.SMTPUser,
		commentSetting.SMTPPassword,
	)
	if err != nil {
		return err
	}
	defer client.Close()
	logUtil.GetLogger().Info("Successfully connected to SMTP server", zap.String("host", commentSetting.SMTPHost), zap.Int("port", commentSetting.SMTPPort))

	// 获取系统设置
	raw, err := s.keyvalueRepository.GetKeyValue(ctx, commonModel.SystemSettingsKey)
	if err != nil {
		return err
	}
	var systemSetting settingModel.SystemSetting
	if err := sonic.Unmarshal([]byte(raw), &systemSetting); err != nil {
		return err
	}

	// 获取评论所属的echo信息
	echo, err := s.repo.GetEchoByID(ctx, comment.EchoID)
	if err != nil {
		return err
	}

	serverName := systemSetting.ServerName
	if serverName == "" {
		serverName = "Ech0"
	}
	serverLogo := systemSetting.ServerLogo
	if serverLogo == "" {
		serverLogo = "/Ech0.svg"
	}
	serverURL := systemSetting.ServerURL
	if serverURL == "" {
		serverURL = "https://ech0.example.com"
	}
	serverLogo = fmt.Sprintf("%s%s", serverURL, serverLogo)
	data := mailUtil.CommentEmailNotificationData{
		Title:     serverName,
		Logo:      serverLogo,
		Host:      serverURL,
		Poster:    echo.Username,
		Commenter: comment.Nickname,
		CommentAt: comment.CreatedAt,
		Content:   comment.Content,
		EchoID:    comment.EchoID,
	}
	// 生成邮件内容
	emailBody, err := mailUtil.GenerateCommentEmailNotification(data)
	if err != nil {
		return err
	}

	getDomain := func(email string) string {
		index := strings.LastIndex(email, "@")
		domain := strings.ToLower(email[index+1:])
		return domain
	}

	// 附加头部字段
	from := commentSetting.SMTPUser
	to := []string{commentSetting.SMTPUser}
	subject := systemSetting.ServerName
	domain := getDomain(commentSetting.SMTPUser)
	email := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Date: "+time.Now().Format(time.RFC1123Z)+"\r\n"+
			"Message-ID: <"+time.Now().Format("20060102150405")+"@%s>\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"\r\n"+
			"%s",
		from, to, subject, domain, emailBody)

	// 发送邮件
	if err := client.SendMail(from, to, strings.NewReader(email)); err != nil {
		return err
	}
	logUtil.GetLogger().Info("Successfully sent email notification", zap.String("from", from), zap.String("to", strings.Join(to, ",")), zap.String("subject", subject))

	return nil
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

func buildDiceBearURL(seed string) string {
	trimmed := strings.TrimSpace(seed)
	if trimmed == "" {
		trimmed = "guest"
	}
	return "https://api.dicebear.com/9.x/fun-emoji/svg?seed=" + url.QueryEscape(trimmed)
}

func deriveCaptchaAPIEndpoint(verifyURL string) string {
	trimmed := strings.TrimSpace(verifyURL)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	path := strings.TrimSuffix(parsed.Path, "/")
	path = strings.TrimSuffix(path, "/siteverify")
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
