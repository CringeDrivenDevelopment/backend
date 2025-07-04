package controller

import (
	"backend/cmd/app"
	"backend/internal/adapters/controller/middlewares"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"io"
	"net/http"
)

type youtubeHandler struct {
	youtubeService *service.YoutubeService
	trackService   *service.TrackService
}

func newYoutubeHandler(app *app.App) *youtubeHandler {
	return &youtubeHandler{
		youtubeService: service.NewYoutubeService(app),
		trackService:   service.NewTrackService(app),
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

	search, err := h.youtubeService.Search(ctx, query, val.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return &struct{ Body []dto.Track }{Body: search}, nil
}

func (h *youtubeHandler) download(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"11" maxLength:"11"`
}) (*struct{}, error) {
	_, ok := ctx.Value(middlewares.UserJwtKey).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}
	id := input.Id

	_, err := h.trackService.GetById(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound(err.Error(), err)
	}

	err = h.youtubeService.Download(ctx, id)
	if err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}

func (h *youtubeHandler) stream(ctx context.Context, input *struct {
	Id   string `path:"id" minLength:"11" maxLength:"11"`
	File string `path:"file"`
}) (*struct {
	Body []byte
}, error) {
	_, ok := ctx.Value(middlewares.UserJwtKey).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	id := input.Id
	_, err := h.trackService.GetById(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound(err.Error(), err)
	}

	stream, err := h.youtubeService.Stream(ctx, id, input.File)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	bytes, err := io.ReadAll(stream)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}
	err = stream.Close()
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	return &struct{ Body []byte }{Body: bytes}, nil
}

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
		Description: "Find tracks by query",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.search)

	huma.Register(router, huma.Operation{
		OperationID: "youtube-dl",
		Path:        "/api/youtube/{id}/dl",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"youtube",
		},
		Summary:     "Download",
		Description: "Download track files",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.download)

	huma.Register(router, huma.Operation{
		OperationID: "youtube-stream",
		Path:        "/api/youtube/{id}/{file}",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"youtube",
		},
		Summary:     "Get file",
		Description: "Get track file",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.stream)
}
