package email

import (
	"astronomer-gin/config"
	"crypto/tls"
	"fmt"

	"gopkg.in/gomail.v2"
)

var Client *EmailClient

// EmailClient 邮件客户端
type EmailClient struct {
	dialer *gomail.Dialer
	config *config.EmailConfig
}

// InitEmail 初始化邮件客户端
func InitEmail(cfg *config.EmailConfig) error {
	if !cfg.Enabled {
		fmt.Println("⚠️  邮件服务未启用")
		return nil
	}

	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.Username, cfg.Password)

	// 配置SSL/TLS
	if cfg.UseSSL {
		dialer.SSL = true
	} else {
		// 使用STARTTLS
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	Client = &EmailClient{
		dialer: dialer,
		config: cfg,
	}

	// 测试连接
	sender, err := dialer.Dial()
	if err != nil {
		return fmt.Errorf("邮件服务器连接失败: %w", err)
	}
	defer sender.Close()

	fmt.Println("✅ 邮件服务初始化成功")
	return nil
}

// SendEmail 发送邮件
func (c *EmailClient) SendEmail(to string, subject string, body string, isHTML bool) error {
	if !c.config.Enabled {
		fmt.Printf("⚠️  邮件服务未启用，跳过发送给 %s 的邮件\n", to)
		return nil
	}

	m := gomail.NewMessage()

	// 设置发件人
	m.SetAddressHeader("From", c.config.Username, c.config.FromName)

	// 设置收件人
	m.SetHeader("To", to)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置邮件正文
	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// 发送邮件
	if err := c.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	fmt.Printf("✅ 邮件已发送给 %s: %s\n", to, subject)
	return nil
}

// SendBatchEmail 批量发送邮件
func (c *EmailClient) SendBatchEmail(recipients []string, subject string, body string, isHTML bool) error {
	if !c.config.Enabled {
		fmt.Println("⚠️  邮件服务未启用，跳过批量发送邮件")
		return nil
	}

	for _, to := range recipients {
		if err := c.SendEmail(to, subject, body, isHTML); err != nil {
			// 记录错误但继续发送其他邮件
			fmt.Printf("❌ 发送邮件给 %s 失败: %v\n", to, err)
		}
	}

	return nil
}

// GetConfig 获取邮件配置
func (c *EmailClient) GetConfig() *config.EmailConfig {
	return c.config
}
