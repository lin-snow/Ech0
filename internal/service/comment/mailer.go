// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wneessen/go-mail"
)

type GoMailSender struct{}

func NewGoMailSender() *GoMailSender {
	return &GoMailSender{}
}

func (s *GoMailSender) Send(ctx context.Context, cfg MailerConfig, msg MailMessage) error {
	host := strings.TrimSpace(cfg.Host)
	to := strings.TrimSpace(msg.To)
	from := strings.TrimSpace(cfg.Sender)
	if from == "" {
		from = strings.TrimSpace(cfg.Username)
	}
	if host == "" || to == "" || from == "" {
		return fmt.Errorf("missing mail configuration")
	}

	port, useSSL, tlsPolicy := resolveSMTPTransport(cfg.Port)
	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithTimeout(10 * time.Second),
		mail.WithTLSPortPolicy(tlsPolicy),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	}
	if useSSL {
		opts = append(opts, mail.WithSSLPort(false))
	}

	if strings.TrimSpace(cfg.Username) != "" {
		opts = append(opts, mail.WithUsername(strings.TrimSpace(cfg.Username)))
	}
	if cfg.Password != "" {
		opts = append(opts, mail.WithPassword(cfg.Password))
	}

	client, err := mail.NewClient(host, opts...)
	if err != nil {
		return err
	}
	if err := client.DialWithContext(ctx); err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	m := mail.NewMsg()
	if err := m.From(from); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(strings.TrimSpace(msg.Subject))
	if strings.TrimSpace(msg.HTMLBody) != "" {
		m.SetBodyString(mail.TypeTextHTML, msg.HTMLBody)
	} else {
		m.SetBodyString(mail.TypeTextPlain, msg.TextBody)
	}
	return client.Send(m)
}

func defaultPort(port int) int {
	if port > 0 {
		return port
	}
	return 587
}

func resolveSMTPTransport(configuredPort int) (port int, sslPort bool, tlsPolicy mail.TLSPolicy) {
	port = defaultPort(configuredPort)

	switch port {
	case 465:
		return port, true, mail.NoTLS
	case 25:
		return port, false, mail.TLSOpportunistic
	default:
		return port, false, mail.TLSMandatory
	}
}
