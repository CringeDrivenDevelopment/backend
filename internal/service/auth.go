package service

import (
	"backend/internal/app"
	"backend/internal/errorz"
	"backend/internal/transport/api/dto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang-jwt/jwt/v5"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type Auth struct {
	secret   string
	botToken string

	expires time.Duration
}

// NewAuthService - создать новый экземпляр сервиса авторизации
func NewAuthService(app *app.App, expires time.Duration) *Auth {
	return &Auth{secret: app.Settings.JwtSecret, botToken: app.Settings.BotToken, expires: expires}
}

// VerifyToken - проверить токен на подлинность
func (s *Auth) VerifyToken(authHeader string) (int64, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return 0, errorz.ErrInvalidToken
	}

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid || token.Method != jwt.SigningMethodHS256 {
		return 0, errorz.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errorz.ErrInvalidToken
	}

	userIDstr, ok := claims["sub"].(string)
	if !ok {
		return 0, errorz.ErrInvalidToken
	}

	userID, err := strconv.ParseInt(userIDstr, 10, 64)
	if err != nil {
		return 0, errorz.ErrInvalidToken
	}

	return userID, nil
}

// GenerateToken - создать новый JWT токен
func (s *Auth) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.expires).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.secret))
}

// ParseInitData - извлечь Telegram ID из Init Data Raw
func (s *Auth) ParseInitData(initDataRaw string) (int64, error) {
	if err := initdata.Validate(initDataRaw, s.botToken, s.expires); err != nil {
		return 0, errorz.ErrInvalidInitData
	}

	initDataValues, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, errorz.ErrInvalidInitData
	}

	initDataUser := initDataValues.Get("user")

	if initDataUser == "" {
		return 0, errorz.ErrInvalidInitData
	}

	user := dto.TelegramData{}
	err = sonic.Unmarshal([]byte(initDataUser), &user)
	if err != nil {
		return 0, errorz.ErrInvalidInitData
	}

	return user.ID, nil
}
