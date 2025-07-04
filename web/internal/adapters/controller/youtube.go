package controller

import (
	"backend/cmd/app"
	"backend/internal/adapters/controller/middlewares"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type youtubeHandler struct {
	youtube *service.YoutubeService
}

func newYoutubeHandler(app *app.App) *youtubeHandler {
	return &youtubeHandler{
		youtube: service.NewYoutubeService(app),
	}
}

func (h *youtubeHandler) search(ctx context.Context, input *struct {
	Query string `query:"query"`
}) (*struct {
	Body []dto.Track
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}
	query := input.Query

	if len(input.Query) < 3 {
		return nil, huma.Error400BadRequest("query too short")
	}

	search, err := h.youtube.Search(ctx, query, val.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return &struct{ Body []dto.Track }{Body: search}, nil
}

/*
func (h *youtubeHandler) download(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"11" maxLength:"11"`
}) (*struct {}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}
	id := input.Id


}
*/

func (h *youtubeHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "youtube-search",
		Path:        "/api/youtube/search",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"youtube",
		},
		Summary:     "Search",
		Description: "Find tracks by query, limit by 5",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.search)
}
