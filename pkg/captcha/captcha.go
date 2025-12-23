package captcha

import (
	"context"
	"image/color"

	"github.com/mojocn/base64Captcha"
)

var Store = base64Captcha.DefaultMemStore

type CaptchaInfo struct {
	CaptchaID string `json:"captchaId"`
	PicPath   string `json:"picPath"`
}

// GetCaptcha 生成验证码
func GetCaptcha(ctx context.Context) (*CaptchaInfo, error) {
	var driver base64Captcha.Driver
	var driverString base64Captcha.DriverString

	//配置验证码信息
	captchaConfig := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          6,
		Source:          "123456789qwertyuiopasdfghjklzxcvbnm",
		BgColor: &color.RGBA{
			R: 3,
			G: 102,
			B: 214,
			A: 125,
		},
		Fonts: []string{"wqy-microhei.ttc"},
	}
	driverString = captchaConfig
	driver = driverString.ConvertFonts()
	captcha := base64Captcha.NewCaptcha(driver, Store)
	lid, lb64s, _, _ := captcha.Generate()

	return &CaptchaInfo{
		CaptchaID: lid,
		PicPath:   lb64s,
	}, nil
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(captchaID, captchaValue string) bool {
	return Store.Verify(captchaID, captchaValue, true)
}
