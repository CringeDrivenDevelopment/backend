package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
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
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}
	query := input.Query

	if len(input.Query) < 3 {
		return nil, huma.Error400BadRequest("query too short")
	}

	search, err := h.youtubeService.Search(ctx, query, val)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return &struct{ Body []dto.Track }{Body: search}, nil
}

func (h *youtubeHandler) download(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"11" maxLength:"11"`
}) (*struct{}, error) {
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

func (h *youtubeHandler) baseStream(ctx context.Context, id *string, file string) (*dto.FileBody, error) {
	if id != nil {
		_, err := h.trackService.GetById(ctx, *id)
		if err != nil {
			return nil, huma.Error404NotFound(err.Error(), err)
		}
	}

	stream, err := h.youtubeService.Stream(ctx, id, file)
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

	return &dto.FileBody{Body: bytes}, nil
}

func (h *youtubeHandler) streamMusic(ctx context.Context, input *struct {
	Id   string `path:"id" minLength:"11" maxLength:"11"`
	File string `path:"file"`
}) (*dto.FileBody, error) {
	if strings.Contains(input.File, "..") {
		return nil, huma.Error400BadRequest("file must not contain '..'")
	}

	if strings.Contains(input.Id, "..") {
		return nil, huma.Error400BadRequest("id must not contain '..'")
	}

	return h.baseStream(ctx, &input.Id, input.File)
}

func (h *youtubeHandler) streamArchive(ctx context.Context, input *struct {
	File string `path:"file"`
}) (*dto.FileBody, error) {
	if strings.Contains(input.File, "..") {
		return nil, huma.Error400BadRequest("file must not contain '..'")
	}

	if !strings.HasSuffix(input.File, ".zip") {
		return nil, huma.Error400BadRequest("file must contain '.zip'")
	}

	return h.baseStream(ctx, nil, input.File)
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
		Method:      http.MethodPost,
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
		OperationID: "youtube-stream-music",
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
		Summary:     "Get music file",
		Description: "Get file by filename",
		/* // TODO: ADD LATER, IMPORTANT, ASAP
		Middlewares: huma.Middlewares{auth},
			Security: []map[string][]string{
				{
					"jwt": []string{},
				},
			},
		*/
	}, h.streamMusic)

	huma.Register(router, huma.Operation{
		OperationID: "youtube-stream-archive",
		Path:        "/api/youtube/{file}",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"youtube",
		},
		Summary:     "Get songs archive",
		Description: "Get file by filename",
	}, h.streamArchive)
}
