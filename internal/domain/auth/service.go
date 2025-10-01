package auth

import (
	"backend/cmd/app"
	"backend/internal/infra/database/queries"
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret string

	expires time.Duration
}

func NewService(app *app.App, expires time.Duration) *Service {
	return &Service{secret: app.Settings.JwtSecret, expires: expires}
}

func (s *Service) VerifyToken(authHeader string) (int64, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return 0, huma.Error401Unauthorized("пустой заголовок Authorization")
	}

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid || token.Method != jwt.SigningMethodHS256 {
		return 0, huma.Error401Unauthorized("токен невалидный или неверный метод подписи")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, huma.Error401Unauthorized("невалидный токен")
	}

	userIDstr, ok := claims["sub"].(string)
	if !ok {
		return 0, huma.Error401Unauthorized("невалидный токен")
	}

	userID, err := strconv.ParseInt(userIDstr, 10, 64)
	if err != nil {
		return 0, huma.Error401Unauthorized("невалидный токен")
	}

	return userID, nil
}

func (s *Service) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.expires).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.secret))
}

func (s *Service) GetUserFromJWT(jwt string, context context.Context, getUser func(context.Context, int64) (queries.User, error)) (queries.User, error) {
	id, errVerify := s.VerifyToken(jwt)
	if errVerify != nil {
		return queries.User{}, errVerify
	}

	user, errGetUser := getUser(context, id)
	if errGetUser != nil {
		return queries.User{}, errGetUser
	}

	return user, nil
}

func ParseInitData(initDataRaw string) (int64, error) {
	initDataValues, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, errors.New("failed to parse url query")
	}

	initDataUser := initDataValues.Get("user")

	if initDataUser == "" {
		return 0, errors.New("telegram user empty")
	}

	user := Data{}
	err = sonic.Unmarshal([]byte(initDataUser), &user)
	if err != nil {
		return 0, errors.New("failed to parse user data")
	}

	return user.ID, nil
}
