package app

import (
	"backend/internal/adapters/config"
	"backend/internal/adapters/controller/validator"
	"backend/internal/domain/utils"
	"context"
	"github.com/bytedance/sonic"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
)

// App is a struct that contains the fiber app, database connection, listen port, validator, logging boolean etc.
type App struct {
	Server    *echo.Echo
	Router    huma.API
	DB        *pgxpool.Pool
	Validator *validator.Validator
	Logger    *zap.Logger

	Settings *config.Settings
}

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

// New is a function that creates a new app struct
func New(logger *zap.Logger) (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	logger.Info("dbconn: " + cfg.DbUrl)

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
	apiCfg.OpenAPI.Servers = append(apiCfg.Servers, &huma.Server{
		URL:         "https://cloud.lxft.tech",
		Description: "PROD",
	}, &huma.Server{
		URL:         "http://localhost:8080",
		Description: "dev",
	})
	router := echo.New()
	router.HideBanner = true
	router.HidePort = true
	api := humaecho.New(router, apiCfg)

	conn, err := utils.NewConnection(context.Background(), cfg.DbUrl)
	if err != nil {
		return nil, err
	}

	requestValidator := validator.New(cfg.BotTokens)

	return &App{
		Server:    router,
		Router:    api,
		DB:        conn,
		Validator: requestValidator,
		Logger:    logger,
		Settings:  cfg,
	}, nil
}

// Start is a function that starts the app
func (a *App) Start() error {
	a.Logger.Info("starting server on :8080")
	if err := a.Server.Start(":8080"); err != nil {
		return err
	}
	return nil
}

func (a *App) Shutdown() error {
	a.Logger.Info("stopping server")
	err := a.Server.Close()
	if err != nil {
		return err
	}

	a.DB.Close()

	return nil
}
