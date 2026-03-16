package mail

import (
	"fmt"
	"time"

	"github.com/matcornic/hermes/v2"
)

type CommentEmailNotificationData struct {
	Title     string
	Logo      string
	Host      string
	Poster    string
	Commenter string
	CommentAt time.Time
	Content   string
	EchoID    string
}

func GenerateCommentEmailNotification(data CommentEmailNotificationData) (string, error) {
	// 动态链接
	echoLink := fmt.Sprintf("%s/echo/%s", data.Host, data.EchoID)

	// 创建 Hermes 实例
	h := hermes.Hermes{
		Product: hermes.Product{
			Name:        data.Title,                              // 网站名称
			Link:        data.Host,                               // 网站链接
			Logo:        data.Logo,                               // 网站Logo
			TroubleText: "如果您无法点击按钮，请将以下网址复制到浏览器中打开：" + echoLink, // 备用文本
		},
	}

	// 定义电子邮件内容
	email := hermes.Email{
		Body: hermes.Body{
			Name: data.Poster, // 动态作者的名字
			Intros: []string{
				"您的动态有了新评论！",
				"以下是评论内容：",
			},
			Dictionary: []hermes.Entry{
				{Key: "评论者", Value: data.Commenter},                                // 评论者的名字
				{Key: "评论时间", Value: data.CommentAt.Format("2006-01-02 15:04:05")}, // 评论时间
				{Key: "评论内容", Value: data.Content},                                 // 评论内容
			},
			Actions: []hermes.Action{
				{
					Instructions: "点击以下按钮查看动态详情：",
					Button: hermes.Button{
						Color: "#22BC66", // 按钮颜色
						Text:  "查看动态",
						Link:  echoLink, // 动态链接
					},
				},
			},
			Outros: []string{
				"如果您不想再接收此类通知，请前往设置页面调整通知偏好。",
			},
		},
	}

	// 生成HTML电子邮件
	emailBody, err := h.GenerateHTML(email)
	if err != nil {
		return "", err
	}

	// // 生成纯文本电子邮件
	// plainTextBody, err := h.GeneratePlainText(email)
	// if err != nil {
	// 	return "", err
	// }

	return emailBody, nil
}
