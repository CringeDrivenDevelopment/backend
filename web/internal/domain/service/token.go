package service

import (
	"backend/internal/adapters/repository"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"strings"
	"time"
)

type TokenService struct {
	secret string

	expires time.Duration
}

func NewTokenService(secret string, expires time.Duration) *TokenService {
	return &TokenService{secret: secret, expires: expires}
}

func (s *TokenService) VerifyToken(authHeader string) (int64, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return 0, errors.New("auth header is empty")
	}

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid || token.Method != jwt.SigningMethodHS256 {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDstr, ok := claims["sub"].(string)
	if !ok {
		return 0, errors.New("invalid token sub")
	}

	userID, err := strconv.ParseInt(userIDstr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user id")
	}

	return userID, nil
}

// GenerateToken is a method to generate a new auth token.
func (s *TokenService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(userID, 10),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.expires).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.secret))
}

func (s *TokenService) GetUserFromJWT(jwt string, context context.Context, getUser func(context.Context, int64) (repository.User, error)) (repository.User, error) {
	id, errVerify := s.VerifyToken(jwt)
	if errVerify != nil {
		return repository.User{}, errVerify
	}

	user, errGetUser := getUser(context, id)
	if errGetUser != nil {
		return repository.User{}, errGetUser
	}

	return user, nil
}
