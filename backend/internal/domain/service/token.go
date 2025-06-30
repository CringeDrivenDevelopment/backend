package service

import (
	"backend/internal/adapters/repository"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
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

func (s *TokenService) verifyToken(authHeader string) (int64, error) {
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return 0, errors.New("auth header is empty")
	}

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(int64)
	if !ok {
		return 0, errors.New("invalid token sub")
	}

	return userID, nil
}

// GenerateToken is a method to generate a new auth token.
func (s *TokenService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.expires).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.secret))
}

func (s *TokenService) GetUserFromJWT(jwt string, context context.Context, getUser func(context.Context, int64) (repository.User, error)) (repository.User, error) {
	id, errVerify := s.verifyToken(jwt)
	if errVerify != nil {
		return repository.User{}, errVerify
	}

	user, errGetUser := getUser(context, id)
	if errGetUser != nil {
		return repository.User{}, errGetUser
	}

	return user, nil
}
