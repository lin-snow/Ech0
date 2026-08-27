// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"fmt"
	stdhtml "html"
	"net/url"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/comment"
)

type notifyContent struct {
	Subject  string
	TextBody string
	HTMLBody string
}

func buildNotifyContent(kind string, comment model.Comment, serverURL string) notifyContent {
	subject := notifySubject(kind, comment.Status)
	eventTitle := notifyEventTitle(kind, comment.Status)
	statusLabel, statusColor, statusBg := notifyStatusStyle(kind, comment.Status)
	createdAt := notifyTime(comment.CreatedAt)
	nickname := fallbackText(comment.Nickname, "匿名用户")
	authorEmail := fallbackText(comment.Email, "未提供")
	contentText := strings.TrimSpace(comment.Content)
	if contentText == "" && kind == "test" {
		contentText = "这是一封来自 Ech0 的评论通知测试邮件，用于验证 SMTP 与模板渲染是否正常。"
	}
	if contentText == "" {
		contentText = "（无正文）"
	}
	echoLink := buildEchoLink(serverURL, comment.EchoID)
	contentHTML := strings.ReplaceAll(stdhtml.EscapeString(contentText), "\n", "<br/>")
	text := buildNotifyText(eventTitle, statusLabel, createdAt, nickname, authorEmail, contentText, echoLink)
	actionHTML := ""
	if echoLink != "" {
		actionHTML = fmt.Sprintf(
			`<div style="margin-top:16px;"><a href="%s" target="_blank" rel="noopener noreferrer" style="display:inline-block;padding:8px 14px;border-radius:0;background:#ffffff;border:1px solid #cbc4b8;color:#5f574a;text-decoration:none;font-size:13px;font-weight:600;">查看 Echo</a></div>`,
			stdhtml.EscapeString(echoLink),
		)
	}
	htmlBody := fmt.Sprintf(
		`<!doctype html><html><body style="margin:0;padding:0;background:#f4f1ec;font-family:'SF Pro Text','PingFang SC','Hiragino Sans GB','Microsoft YaHei',-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Arial,sans-serif;color:#3a3329;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="padding:28px 14px;">
  <tr><td align="center">
    <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:640px;background:#ffffff;border:1px solid #e6dfd4;border-radius:0;overflow:hidden;">
      <tr><td style="padding:16px 20px;border-bottom:1px solid #ebe5db;">
        <div style="font-size:16px;font-weight:700;color:#3a3329;">Ech0 评论通知</div>
      </td></tr>
      <tr><td style="padding:20px;">
        <div style="display:inline-block;padding:3px 10px;border-radius:0;font-size:12px;font-weight:600;color:%s;background:%s;border:1px solid %s;">%s</div>
        <div style="margin-top:12px;font-size:18px;font-weight:700;color:#3a3329;">%s</div>
        <div style="margin-top:14px;padding:14px;border:1px solid #e8e2d8;border-radius:0;background:#fffcf8;line-height:1.7;font-size:14px;color:#4f473b;word-break:break-word;">%s</div>
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="margin-top:10px;border-collapse:collapse;background:#faf7f2;border:1px solid #e8e2d8;border-radius:0;">
          <tr><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#8b8377;width:72px;">动作</td><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#5f574a;">%s</td></tr>
          <tr><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#8b8377;">时间</td><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#5f574a;">%s</td></tr>
          <tr><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#8b8377;">昵称</td><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#5f574a;">%s</td></tr>
          <tr><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#8b8377;">邮箱</td><td style="padding:7px 10px;font-size:12px;line-height:1.45;color:#5f574a;">%s</td></tr>
        </table>
        %s
        <div style="margin-top:14px;font-size:12px;line-height:1.6;color:#958d80;">此邮件由 Ech0 评论系统自动发送。</div>
      </td></tr>
    </table>
  </td></tr>
</table>
</body></html>`,
		statusColor,
		statusBg,
		statusBg,
		stdhtml.EscapeString(statusLabel),
		stdhtml.EscapeString(eventTitle),
		contentHTML,
		stdhtml.EscapeString(eventTitle),
		stdhtml.EscapeString(createdAt),
		stdhtml.EscapeString(nickname),
		stdhtml.EscapeString(authorEmail),
		actionHTML,
	)
	return notifyContent{
		Subject:  subject,
		TextBody: text,
		HTMLBody: htmlBody,
	}
}

func buildNotifyText(
	eventTitle string,
	statusLabel string,
	createdAt string,
	nickname string,
	authorEmail string,
	contentText string,
	echoLink string,
) string {
	text := fmt.Sprintf(
		"Ech0 评论通知\n\n动作: %s\n状态: %s\n时间: %s\n作者: %s <%s>\n\n评论正文:\n%s",
		eventTitle,
		statusLabel,
		createdAt,
		nickname,
		authorEmail,
		contentText,
	)
	if echoLink != "" {
		text += fmt.Sprintf("\n\n查看 Echo:\n%s", echoLink)
	}
	return text
}

func buildEchoLink(serverURL string, echoID string) string {
	base := strings.TrimSpace(serverURL)
	id := strings.TrimSpace(echoID)
	if base == "" || id == "" {
		return ""
	}
	base = strings.TrimSuffix(base, "/")
	return base + "/echo/" + url.PathEscape(id)
}

func fallbackText(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func notifySubject(kind string, status model.Status) string {
	prefix := "[Ech0]"
	switch kind {
	case "created":
		return prefix + " 新评论待审核"
	case "reply":
		return prefix + " 有人回复了你的评论"
	case "status":
		if status == model.StatusApproved {
			return prefix + " 您的评论已通过审核"
		}
		if status == model.StatusRejected {
			return prefix + " 您的评论未通过审核"
		}
		return prefix + " 评论状态已更新"
	case "hot":
		return prefix + " 您的评论被标为精选"
	default:
		return prefix + " 邮件通知测试"
	}
}

func notifyEventTitle(kind string, status model.Status) string {
	switch kind {
	case "created":
		return "有新评论待处理"
	case "reply":
		return "有人回复了你的评论"
	case "status":
		if status == model.StatusApproved {
			return "评论已审核通过"
		}
		if status == model.StatusRejected {
			return "评论已审核拒绝"
		}
		return "评论状态已更新"
	case "hot":
		return "评论已被设为 Hot"
	default:
		return "评论通知测试邮件"
	}
}

func notifyStatusStyle(kind string, status model.Status) (label, color, bg string) {
	switch kind {
	case "hot":
		return "HOT", "#7c3aed", "#f3e8ff"
	case "reply":
		return "回复", "#0369a1", "#e0f2fe"
	case "status":
		if status == model.StatusApproved {
			return "已通过", "#059669", "#ecfdf5"
		}
		if status == model.StatusRejected {
			return "已拒绝", "#d97706", "#fff7ed"
		}
	}
	if status == model.StatusPending {
		return "待审核", "#0369a1", "#e0f2fe"
	}
	if status == model.StatusApproved {
		return "已通过", "#059669", "#ecfdf5"
	}
	if status == model.StatusRejected {
		return "已拒绝", "#d97706", "#fff7ed"
	}
	return "通知", "#6b6458", "#f5f1ea"
}

func notifyTime(ts int64) string {
	if ts == 0 {
		ts = time.Now().Unix()
	}
	return time.Unix(ts, 0).Local().Format("2006-01-02 15:04:05")
}
