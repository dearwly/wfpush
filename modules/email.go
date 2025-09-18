package modules

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

// 修改函数签名以接收配置
func Send_email(body string, cfg *Config) {
	// 从配置中获取邮件服务器信息
	smtpServer := cfg.Email.SMTPServer
	port := strconv.Itoa(cfg.Email.Port)
	sender := cfg.Email.Sender
	password := cfg.Email.Password
	recipients := cfg.Email.Recipients

	if len(recipients) == 0 {
		Log("No recipients configured in config.yml. Email not sent.", WARNING)
		return
	}

	// 邮件主题
	subject := "Warframe 裂缝订阅通知"

	// 构建邮件内容
	// To 头部可以是逗号分隔的列表
	message := []byte("Subject: " + subject + "\r\n" +
		"From: " + sender + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"\r\n" +
		body)

	// 使用 SSL/TLS 连接到邮件服务器
	conn, err := tls.Dial("tcp", smtpServer+":"+port, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	})
	if err != nil {
		Log(fmt.Sprintf("Failed to connect to the server: %v", err), ERROR)
		return
	}

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		Log(fmt.Sprintf("Failed to create SMTP client: %v", err), ERROR)
		return
	}

	// 认证
	auth := smtp.PlainAuth("", sender, password, smtpServer)
	if err := client.Auth(auth); err != nil {
		Log(fmt.Sprintf("Authentication failed: %v", err), ERROR)
		return
	}

	// 设置发件人
	if err := client.Mail(sender); err != nil {
		Log(fmt.Sprintf("Failed to set mail from: %v", err), ERROR)
		return
	}

	// 遍历并设置所有收件人
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			Log(fmt.Sprintf("Failed to set recipient %s: %v", recipient, err), WARNING)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		Log(fmt.Sprintf("Failed to create writer: %v", err), ERROR)
		return
	}
	_, err = writer.Write(message)
	if err != nil {
		Log(fmt.Sprintf("Failed to write message: %v", err), ERROR)
		return
	}
	err = writer.Close()
	if err != nil {
		Log(fmt.Sprintf("Failed to close writer: %v", err), ERROR)
		return
	}

	// 关闭连接
	client.Quit()

	Log(fmt.Sprintf("Email sent successfully to: %s", strings.Join(recipients, ", ")), INFO)
}
