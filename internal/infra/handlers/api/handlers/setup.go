package handlers

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Setup - добавить маршрут до эндпоинта
func (h *User) Setup(router huma.API) {
	huma.Register(router, huma.Operation{
		OperationID: "auth",
		Path:        "/api/auth",
		Method:      http.MethodPost,
		Errors: []int{
			401,
			422,
			500,
		},
		Tags: []string{
			"auth",
		},
		Summary:     "Login",
		Description: "Получить токен для взаимодействия. Нуждается в InitDataRaw строке из Telegram Mini App. Действует 1 час",
	}, h.login)
}

// Setup - добавить маршрут до эндпоинтов
func (h *Playlist) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "playlist-by-id",
		Path:        "/api/playlists/{id}",
		Method:      http.MethodGet,
		Errors: []int{
			401,
			404,
			422,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Get by ID",
		Description: "Получить плейлист по ID. Для получения требуется, чтобы у юзера были права на просмотр плейлиста. При получении вернёт массив треков",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.getById)

	huma.Register(router, huma.Operation{
		OperationID: "playlist-all",
		Path:        "/api/playlists",
		Method:      http.MethodGet,
		Errors: []int{
			401,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "All",
		Description: "Получить весь список плейлистов. Вернёт только те плейлисты, к которым у пользователя есть доступ. При получении не вернёт массив треков",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.getAll)
}

func (h *Track) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "track-search",
		Path:        "/api/search",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			422,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Search",
		Description: "Найти трек по запросу. Поиск по Youtube Music",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.search)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-submit",
		Path:        "/api/playlists/{playlist_id}/{track_id}/submit",
		Method:      http.MethodPost,
		Errors: []int{
			401,
			404,
			422,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Playlist Submit",
		Description: "Добавить трек в плейлист, если юзер есть в плейлисте. Если у юзера права админа, то трек добавляется в разрешённые, иначе на модерацию",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.submit)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-unapprove",
		Path:        "/api/playlists/{playlist_id}/{track_id}/unapprove",
		Method:      http.MethodDelete,
		Errors: []int{
			401,
			404,
			422,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Playlist Unapprove",
		Description: "Убрать трек из разрешённых. У юзера должны быть права админа",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.unapprove)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-approve",
		Path:        "/api/playlists/{playlist_id}/{track_id}/approve",
		Method:      http.MethodPatch,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Playlist Approve",
		Description: "Добавить трек в разрешённые. У юзера должны быть права админа",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.approve)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-decline",
		Path:        "/api/playlists/{playlist_id}/{track_id}/decline",
		Method:      http.MethodDelete,
		Errors: []int{
			401,
			403,
			404,
			422,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Playlist Decline",
		Description: "Удалить трек из кандидатов в плейлист. У юзера должны быть права админа",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.decline)
}
