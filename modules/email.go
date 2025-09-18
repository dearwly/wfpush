package modules

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

// HTMLTemplate 修改：在显示地点时调用 TranslateNode
const HTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif; margin: 20px; color: #333; background-color: #f9f9f9; }
  .container { border: 1px solid #e0e0e0; padding: 25px; border-radius: 10px; max-width: 700px; margin: auto; background-color: #fff; box-shadow: 0 4px 8px rgba(0,0,0,0.05); }
  h1 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; }
  .fissure-table { width: 100%; border-collapse: collapse; margin-top: 25px; }
  .fissure-table th, .fissure-table td { border: 1px solid #e0e0e0; padding: 12px; text-align: left; }
  .fissure-table th { background-color: #3498db; color: #fff; font-weight: bold; }
  .fissure-table tr:nth-child(even) { background-color: #f2f8fc; }
  .fissure-table tr:hover { background-color: #eaf4fb; }
  .fissure-type { font-weight: bold; }
  .steel-path { color: #c0392b; font-weight: bold; }
  .void-storm { color: #8e44ad; font-weight: bold; }
  .footer { font-size: 12px; color: #95a5a6; margin-top: 25px; text-align: center; }
</style>
</head>
<body>
  <div class="container">
    <h1>Warframe 裂缝订阅通知</h1>
    <p>您好！您订阅的以下新裂缝任务已出现：</p>
    <table class="fissure-table">
      <thead>
        <tr>
          <th>任务</th>
          <th>地点</th>
          <th>纪元</th>
          <th>阵营</th>
          <th>剩余时间</th>
        </tr>
      </thead>
      <tbody>
        {{range .}}
        <tr>
          <td>
            <span class="fissure-type">
              {{if .IsHard}}<span class="steel-path">钢铁之路</span> {{end}}
              {{if .IsStorm}}<span class="void-storm">虚空风暴</span> {{end}}
              {{.MissionType | TranslateMissionType}}
            </span>
          </td>
          <td>{{.Node | TranslateNode}}</td>
          <td>{{.Tier | TranslateTier}}</td>
          <td>{{.EnemyKey | TranslateFaction}}</td>
          <td>{{.GetETA}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
    <p class="footer">
      由 Warframe 小助手自动发送
    </p>
  </div>
</body>
</html>`

// Send_email 函数保持不变
func Send_email(htmlBody string, cfg *Config) {
	// ... 省略未改变的代码 ...
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

	headers := "From: " + sender + "\r\n" +
		"To: " + strings.Join(recipients, ",") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"

	message := []byte(headers + "\r\n" + htmlBody)

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
