package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/validator"
	"backend/internal/domain/auth"
	"backend/internal/domain/user"
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type userHandler struct {
	userService  *user.Service
	tokenService *auth.Service
	validator    *validator.Validator
}

func newUserHandler(app *app.App) *userHandler {
	userService := user.NewService(app)
	tokenService := auth.NewService(app, time.Hour)

	return &userHandler{
		userService:  userService,
		tokenService: tokenService,
		validator:    app.Validator,
	}
}

func (h *userHandler) login(ctx context.Context, input *auth.InputStruct) (*auth.OutputStruct, error) {
	userDTO := input.Body

	if errValidate := h.validator.ValidateData(userDTO); errValidate != nil {
		return nil, huma.Error400BadRequest(errValidate.Error(), errValidate)
	}

	id, tgErr := auth.ParseInitData(userDTO.InitDataRaw)
	if tgErr != nil {
		return nil, huma.Error400BadRequest(tgErr.Error(), tgErr)
	}

	errFetch := h.userService.GetByID(ctx, id)
	if errFetch != nil {
		var createErr error

		createErr = h.userService.Create(ctx, id)

		if createErr != nil {
			return nil, huma.Error400BadRequest(createErr.Error(), createErr)
		}
	}

	token, tokenErr := h.tokenService.GenerateToken(id)
	if tokenErr != nil || token == "" {
		return nil, huma.Error500InternalServerError("failed to generate auth token")
	}

	tokenData := auth.TokenDto{
		Token: token,
	}

	return &auth.OutputStruct{Body: tokenData}, nil
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
