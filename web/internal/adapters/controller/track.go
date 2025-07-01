package controller

import (
	"backend/cmd/app"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type tracksHandler struct {
	youtube *service.YoutubeService
}

func newTracksHandler(app *app.App) *tracksHandler {
	return &tracksHandler{
		youtube: service.NewYoutubeService(app),
	}
}

func (h *tracksHandler) search(ctx context.Context, input *dto.SearchInput) (*struct {
	Body []dto.Track
}, error) {
	query := input.Query

	if len(input.Query) < 3 {
		return nil, huma.Error400BadRequest("query too short")
	}

	search, err := h.youtube.Search(ctx, query)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return &struct{ Body []dto.Track }{Body: search}, nil
}

func (h *tracksHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "tracks-search",
		Path:        "/api/tracks",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			500,
		},
		Tags: []string{
			"tracks",
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
