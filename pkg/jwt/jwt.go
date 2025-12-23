package jwt

import (
	"astronomer-gin/config"
	"errors"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type LoginClaims struct {
	Phone                string `json:"phone"`   // 保留用于兼容
	UserID               string `json:"user_id"` // 新增：直接存储UUID
	jwt.RegisteredClaims        //内嵌标准的声明
}

var (
	letters = []rune("0123456789qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM")
)

func RandStr(strLen int) string {
	randBytes := make([]rune, strLen)
	for i := range randBytes {
		randBytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(randBytes)
}

// Sign 生成签名token
func Sign(phone, userID string) (string, error) {
	cfg := config.GlobalConfig.JWT
	claim := LoginClaims{
		Phone:  phone,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Auth_Server",                                                                  //jwt签发者
			Subject:   "Auth",                                                                         //jwt面向的用户
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.ExpireHours))), // 根据配置设置过期时间
			NotBefore: jwt.NewNumericDate(time.Now().Add(time.Second)),                                //最早使用时长一秒后
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                 //签发时间即刻
			ID:        RandStr(10),                                                                    //随机十位字符为token
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claim).SignedString([]byte(cfg.SecretKey))
	return token, err
}

// Verify 检验token是否正确
func Verify(tokenString string) (*LoginClaims, error) {
	cfg := config.GlobalConfig.JWT
	token, err := jwt.ParseWithClaims(tokenString, &LoginClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil //返回签名密钥
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("claim invalid")
	}

	claims, ok := token.Claims.(*LoginClaims)
	if !ok {
		return nil, errors.New("invalid claim type")
	}

	return claims, nil
}

// GetLoginPhone 获取token中包含的登录用户手机号
func GetLoginPhone(token string) string {
	claims, err := Verify(token)
	if err != nil {
		return ""
	}
	return claims.Phone
}

// GetLoginUserID 获取token中包含的登录用户ID(UUID)
func GetLoginUserID(token string) string {
	claims, err := Verify(token)
	if err != nil {
		return ""
	}
	return claims.UserID
}
