package modules

import (
	"crypto/tls"
	"net/smtp"
)

func Send_email(body string) {
	// 邮件服务器配置
	smtpServer := "smtp.163.com"
	port := "465"                           // 使用 465 端口进行 SSL/TLS 加密连接
	sender := "15906376758@163.com"         // 发件人邮箱
	password := "wly3369874152"             // 发件人邮箱的密码或App专用密码
	recipient := "dearwangliyu@outlook.com" // 收件人邮箱

	// 邮件主题
	subject := "Warframe 小助手"

	// 构建邮件内容，确保头部和正文格式正确
	message := []byte("Subject: " + subject + "\r\n" +
		"From: " + sender + "\r\n" +
		"To: " + recipient + "\r\n" +
		"\r\n" + // 邮件头部和正文之间的空行
		body)

	// 使用 SSL/TLS 连接到邮件服务器
	conn, err := tls.Dial("tcp", smtpServer+":"+port, &tls.Config{
		InsecureSkipVerify: true, // 如果证书无法验证，跳过验证（不推荐在生产环境中使用）
		ServerName:         smtpServer,
	})
	if err != nil {
		Log("Failed to connect to the server:", ERROR)
		return
	}

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		Log("Failed to create SMTP client:", ERROR)
		return
	}

	// 认证
	auth := smtp.PlainAuth("", sender, password, smtpServer)
	if err := client.Auth(auth); err != nil {
		Log("Authentication failed:", ERROR)
		return
	}

	// 设置发件人和收件人
	if err := client.Mail(sender); err != nil {
		Log("Failed to set mail from:", ERROR)
		return
	}
	if err := client.Rcpt(recipient); err != nil {
		Log("Failed to set recipient:", ERROR)
		return
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		Log("Failed to create writer:", ERROR)
		return
	}
	_, err = writer.Write(message)
	if err != nil {
		Log("Failed to write message:", ERROR)
		return
	}
	err = writer.Close()
	if err != nil {
		Log("Failed to close writer:", ERROR)
		return
	}

	// 关闭连接
	client.Quit()

	Log("Email sent successfully!", INFO)
}
