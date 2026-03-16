package mail

import (
	"crypto/tls"
	"fmt"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

func GetSMTPClient(host string, port int, username, password string) (*smtp.Client, error) {
	// TLS配置
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	// 连接到SMTP服务器
	client, err := smtp.DialStartTLS(fmt.Sprintf("%s:%d", host, port), tlsConfig)
	if err != nil {
		client, err = smtp.DialTLS(fmt.Sprintf("%s:%d", host, port), tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
	}

	// 测试认证
	auth := sasl.NewPlainClient("", username, password)
	if err := client.Auth(auth); err != nil {
		return nil, fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// 测试连接
	if err := client.Noop(); err != nil {
		return nil, fmt.Errorf("SMTP connection unavailable: %w", err)
	}

	return client, nil
}
