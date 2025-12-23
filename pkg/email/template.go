package email

import (
	"fmt"
	"time"
)

// EmailTemplate 邮件模板接口
type EmailTemplate interface {
	GetSubject() string
	GetBody() string
}

// ============================= 欢迎邮件模板 =============================

// WelcomeEmailData 欢迎邮件数据
type WelcomeEmailData struct {
	Username string
	Email    string
}

// GetSubject 获取欢迎邮件主题
func (d *WelcomeEmailData) GetSubject() string {
	if Client != nil && Client.config.Templates.WelcomeSubject != "" {
		return Client.config.Templates.WelcomeSubject
	}
	return "欢迎加入Astronomer博客平台"
}

// GetBody 获取欢迎邮件正文（HTML格式）
func (d *WelcomeEmailData) GetBody() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 欢迎加入 Astronomer</h1>
        </div>
        <div class="content">
            <h2>你好，%s！</h2>
            <p>感谢你注册 <strong>Astronomer博客平台</strong>！</p>
            <p>我们很高兴你加入我们的社区。在这里，你可以：</p>
            <ul>
                <li>📝 发表自己的技术博客和见解</li>
                <li>💬 与其他开发者交流和评论</li>
                <li>❤️ 收藏和关注喜欢的作者</li>
                <li>🔔 接收实时的互动通知</li>
            </ul>
            <p>立即开始你的创作之旅吧！</p>
            <a href="http://localhost:8080" class="button">访问平台</a>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>© 2025 Astronomer博客平台. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, d.Username)
}

// ============================= 评论通知邮件模板 =============================

// CommentNotificationData 评论通知邮件数据
type CommentNotificationData struct {
	Username       string // 收件人用户名
	CommenterName  string // 评论者用户名
	ArticleTitle   string // 文章标题
	CommentContent string // 评论内容
	ArticleID      uint64 // 文章ID
}

// GetSubject 获取评论通知邮件主题
func (d *CommentNotificationData) GetSubject() string {
	if Client != nil && Client.config.Templates.CommentSubject != "" {
		return Client.config.Templates.CommentSubject
	}
	return "您有新的评论通知"
}

// GetBody 获取评论通知邮件正文（HTML格式）
func (d *CommentNotificationData) GetBody() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .comment-box { background: white; padding: 20px; border-left: 4px solid #667eea; margin: 20px 0; border-radius: 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💬 新评论通知</h1>
        </div>
        <div class="content">
            <h2>你好，%s！</h2>
            <p><strong>%s</strong> 评论了你的文章 《%s》：</p>
            <div class="comment-box">
                <p><em>"%s"</em></p>
            </div>
            <p>点击下面的按钮查看详情并回复：</p>
            <a href="http://localhost:8080/api/v1/blog/%d" class="button">查看评论</a>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>© 2025 Astronomer博客平台. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, d.Username, d.CommenterName, d.ArticleTitle, d.CommentContent, d.ArticleID)
}

// ============================= 点赞通知邮件模板 =============================

// LikeNotificationData 点赞通知邮件数据
type LikeNotificationData struct {
	Username     string // 收件人用户名
	LikerName    string // 点赞者用户名
	ArticleTitle string // 文章标题
	ArticleID    uint64 // 文章ID
	LikeCount    int    // 总点赞数
}

// GetSubject 获取点赞通知邮件主题
func (d *LikeNotificationData) GetSubject() string {
	if Client != nil && Client.config.Templates.LikeSubject != "" {
		return Client.config.Templates.LikeSubject
	}
	return "您的文章收到了新的点赞"
}

// GetBody 获取点赞通知邮件正文（HTML格式）
func (d *LikeNotificationData) GetBody() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .stats-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; text-align: center; }
        .button { display: inline-block; padding: 12px 30px; background: #f5576c; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>❤️ 新点赞通知</h1>
        </div>
        <div class="content">
            <h2>你好，%s！</h2>
            <p><strong>%s</strong> 点赞了你的文章 《%s》！</p>
            <div class="stats-box">
                <h3>👍 当前点赞数：%d</h3>
                <p>你的文章正在获得越来越多的认可！</p>
            </div>
            <a href="http://localhost:8080/api/v1/blog/%d" class="button">查看文章</a>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>© 2025 Astronomer博客平台. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, d.Username, d.LikerName, d.ArticleTitle, d.LikeCount, d.ArticleID)
}

// ============================= 关注通知邮件模板 =============================

// FollowNotificationData 关注通知邮件数据
type FollowNotificationData struct {
	Username      string    // 收件人用户名
	FollowerName  string    // 关注者用户名
	FollowerBio   string    // 关注者简介
	FollowTime    time.Time // 关注时间
	FollowerCount int       // 总粉丝数
}

// GetSubject 获取关注通知邮件主题
func (d *FollowNotificationData) GetSubject() string {
	if Client != nil && Client.config.Templates.FollowSubject != "" {
		return Client.config.Templates.FollowSubject
	}
	return "您有新的粉丝"
}

// GetBody 获取关注通知邮件正文（HTML格式）
func (d *FollowNotificationData) GetBody() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #a8edea 0%%, #fed6e3 100%%); color: #333; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .profile-box { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .stats-box { background: white; padding: 15px; border-radius: 5px; margin: 20px 0; text-align: center; }
        .button { display: inline-block; padding: 12px 30px; background: #4db8ff; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👤 新粉丝通知</h1>
        </div>
        <div class="content">
            <h2>你好，%s！</h2>
            <p><strong>%s</strong> 关注了你！</p>
            <div class="profile-box">
                <h3>关于 %s</h3>
                <p>%s</p>
                <p><small>关注时间：%s</small></p>
            </div>
            <div class="stats-box">
                <h3>🎉 当前粉丝数：%d</h3>
            </div>
            <a href="http://localhost:8080" class="button">查看主页</a>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿直接回复。</p>
            <p>© 2025 Astronomer博客平台. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, d.Username, d.FollowerName, d.FollowerName, d.FollowerBio, d.FollowTime.Format("2006-01-02 15:04:05"), d.FollowerCount)
}

// ============================= 辅助函数 =============================

// SendWelcomeEmail 发送欢迎邮件
func SendWelcomeEmail(email string, username string) error {
	if Client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}

	data := &WelcomeEmailData{
		Username: username,
		Email:    email,
	}

	return Client.SendEmail(email, data.GetSubject(), data.GetBody(), true)
}

// SendCommentNotification 发送评论通知邮件
func SendCommentNotification(email string, data *CommentNotificationData) error {
	if Client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}

	return Client.SendEmail(email, data.GetSubject(), data.GetBody(), true)
}

// SendLikeNotification 发送点赞通知邮件
func SendLikeNotification(email string, data *LikeNotificationData) error {
	if Client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}

	return Client.SendEmail(email, data.GetSubject(), data.GetBody(), true)
}

// SendFollowNotification 发送关注通知邮件
func SendFollowNotification(email string, data *FollowNotificationData) error {
	if Client == nil {
		return fmt.Errorf("邮件客户端未初始化")
	}

	return Client.SendEmail(email, data.GetSubject(), data.GetBody(), true)
}
