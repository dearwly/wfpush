package modules

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

// 修改函数以支持HTML邮件
func Send_email(htmlBody string, cfg *Config) {
	smtpServer := cfg.Email.SMTPServer
	port := strconv.Itoa(cfg.Email.Port)
	sender := cfg.Email.Sender
	password := cfg.Email.Password
	recipients := cfg.Email.Recipients

	if len(recipients) == 0 {
		Log("收件人列表为空，邮件未发送。", WARNING)
		return
	}

	subject := "Warframe 裂缝订阅通知"

	// 构建邮件头，关键是设置 Content-Type 为 text/html
	headers := "From: " + sender + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"

	message := []byte(headers + "\r\n" + htmlBody)

	// ... 后续的 SMTP 连接和发送逻辑保持不变 ...
	conn, err := tls.Dial("tcp", smtpServer+":"+port, &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	})
	if err != nil {
		Log(fmt.Sprintf("连接邮件服务器失败: %v", err), ERROR)
		return
	}

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		Log(fmt.Sprintf("创建 SMTP 客户端失败: %v", err), ERROR)
		return
	}

	auth := smtp.PlainAuth("", sender, password, smtpServer)
	if err := client.Auth(auth); err != nil {
		Log(fmt.Sprintf("邮箱认证失败: %v", err), ERROR)
		return
	}

	if err := client.Mail(sender); err != nil {
		Log(fmt.Sprintf("设置发件人失败: %v", err), ERROR)
		return
	}

	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			Log(fmt.Sprintf("设置收件人 %s 失败: %v", recipient, err), WARNING)
		}
	}

	writer, err := client.Data()
	if err != nil {
		Log(fmt.Sprintf("创建写入流失败: %v", err), ERROR)
		return
	}
	_, err = writer.Write(message)
	if err != nil {
		Log(fmt.Sprintf("写入邮件内容失败: %v", err), ERROR)
		return
	}
	err = writer.Close()
	if err != nil {
		Log(fmt.Sprintf("关闭写入流失败: %v", err), ERROR)
		return
	}
	client.Quit()
	Log(fmt.Sprintf("邮件成功发送至: %s", strings.Join(recipients, ", ")), INFO)
}
