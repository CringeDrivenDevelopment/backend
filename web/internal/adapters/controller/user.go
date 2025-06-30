package controller

import (
	"backend/cmd/app"
	"backend/internal/adapters/controller/validator"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"backend/internal/domain/utils"
	"context"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
	"time"
)

type userHandler struct {
	userService  *service.UserService
	tokenService *service.TokenService
	validator    *validator.Validator
}

func newUserHandler(app *app.App) *userHandler {
	userService := service.NewUserService(app.DB)
	tokenService := service.NewTokenService(app.Settings.JwtSecret, time.Hour)

	return &userHandler{
		userService:  userService,
		tokenService: tokenService,
		validator:    app.Validator,
	}
}

func (h *userHandler) login(ctx context.Context, input *dto.UserLoginInput) (*struct {
	Body dto.Token
}, error) {
	userDTO := input.Body

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return nil, huma.Error400BadRequest(errValidate.Error(), errValidate)
	}

	telegramData, tgErr := utils.ParseInitData(userDTO.InitDataRaw)
	if tgErr != nil {
		return nil, huma.Error400BadRequest(tgErr.Error(), tgErr)
	}

	_, errFetch := h.userService.GetByID(ctx, telegramData.ID)
	if errFetch != nil {
		var createErr error

		name := telegramData.Username
		if name == "" {
			name = fmt.Sprintf("user%d", telegramData.ID)
		}

		createErr = h.userService.Create(ctx, repository.CreateUserParams{
			ID:   telegramData.ID,
			Name: name,
		})

		if createErr != nil {
			return nil, huma.Error400BadRequest(createErr.Error(), createErr)
		}
	}

	token, tokenErr := h.tokenService.GenerateToken(telegramData.ID)
	if tokenErr != nil || token == "" {
		return nil, huma.Error500InternalServerError("failed to generate auth token")
	}

	return &struct{ Body dto.Token }{Body: dto.Token{Token: token}}, nil
}

func (h *userHandler) Setup(router huma.API) {
	huma.Register(router, huma.Operation{
		OperationID: "login",
		Path:        "/api/user/login",
		Method:      http.MethodPost,
		Errors: []int{
			400,
			500,
		},
		Tags: []string{
			"user",
		},
		Summary:     "Login",
		Description: "Login to app with telegram webapp init data. Returns his token",
	}, h.login)
}
