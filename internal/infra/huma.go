package infra

import (
	"io"

	"github.com/bytedance/sonic"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
)

func sonicFormat() huma.Format {
	return huma.Format{
		Marshal: func(w io.Writer, v any) error {
			data, err := sonic.Marshal(v)
			if err != nil {
				return err
			}
			_, err = w.Write(data)
			return err
		},
		Unmarshal: sonic.Unmarshal,
	}
}

func defaultApiConfig() huma.Config {
	apiCfg := huma.DefaultConfig("backend", "v1.0.0")
	apiCfg.SchemasPath = "/docs#/schemas"

	apiCfg.Formats = map[string]huma.Format{
		"json":             sonicFormat(),
		"application/json": sonicFormat(),
	}

	apiCfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"jwt": {
			Type:         "http",
			BearerFormat: "JWT",
			Scheme:       "Bearer",
		},
	}

	apiCfg.Servers = append(apiCfg.Servers, &huma.Server{
		URL:         "http://localhost:8080",
		Description: "dev",
	})

	return apiCfg
}

func NewHuma(echo *echo.Echo) huma.API {
	api := humaecho.New(echo, defaultApiConfig())

	return api
}
