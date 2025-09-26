package utils

import (
	"backend/internal/domain/dto"
	"encoding/json"
	"errors"
	"net/url"
)

func ParseInitData(initDataRaw string) (*dto.UserReturn, error) {
	initDataValues, err := url.ParseQuery(initDataRaw)

	if err != nil {
		return nil, err
	}

	initDataUser := initDataValues.Get("user")

	if initDataUser == "" {
		return nil, errors.New("telegram user empty")
	}

	user := dto.UserReturn{}
	err = json.Unmarshal([]byte(initDataUser), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
