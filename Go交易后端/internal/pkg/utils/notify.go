package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"exchange-go/config"
	"exchange-go/internal/pkg/logger"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// SMSConfig 短信配置
type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}

// SendEmail 发送邮件
func SendEmail(to, subject, body string) error {
	// TODO: 从配置读取邮件配置
	cfg := &EmailConfig{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "noreply@example.com",
		Password: "password",
		From:     "noreply@example.com",
		FromName: "Exchange",
	}

	// 构建邮件内容
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.From)
	headers["To"] = to
	headers["Subject"] = subject
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         cfg.Host,
	}

	// 连接服务器
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), tlsConfig)
	if err != nil {
		return fmt.Errorf("连接邮件服务器失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	// 发送邮件
	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取Writer失败: %w", err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("关闭Writer失败: %w", err)
	}

	logger.Infof("邮件发送成功: %s", to)
	return nil
}

// SendVerifyCodeEmail 发送验证码邮件
func SendVerifyCodeEmail(to, code string) error {
	subject := "验证码"
	body := fmt.Sprintf(`
		<div style="padding: 20px; font-family: Arial, sans-serif;">
			<h2>验证码</h2>
			<p>您的验证码是：<strong style="font-size: 24px; color: #1890ff;">%s</strong></p>
			<p>验证码有效期为5分钟，请勿泄露给他人。</p>
			<p style="color: #999; font-size: 12px;">如非本人操作，请忽略此邮件。</p>
		</div>
	`, code)

	// 开发模式不实际发送
	if config.GlobalConfig.App.Mode == "debug" {
		logger.Infof("[开发模式] 模拟发送邮件验证码到 %s: %s", to, code)
		return nil
	}

	return SendEmail(to, subject, body)
}

// SendSMS 发送短信 (阿里云SMS)
func SendSMS(phone, code string) error {
	// TODO: 实现阿里云短信发送
	// 开发模式不实际发送
	if config.GlobalConfig.App.Mode == "debug" {
		logger.Infof("[开发模式] 模拟发送短信验证码到 %s: %s", phone, code)
		return nil
	}

	// 实际发送逻辑
	// ...

	return nil
}

// IsEmail 判断是否是邮箱
func IsEmail(str string) bool {
	return strings.Contains(str, "@") && strings.Contains(str, ".")
}

// IsMobile 判断是否是手机号
func IsMobile(str string) bool {
	if len(str) < 10 || len(str) > 15 {
		return false
	}
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// MaskPhone 手机号脱敏
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskEmail 邮箱脱敏
func MaskEmail(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex < 2 {
		return email
	}
	return email[:2] + "****" + email[atIndex:]
}
