package app

import (
	"context"
	"os"

	"github.com/4udiwe/avito-pr-service/config"
	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/database"
	repo_pr "github.com/4udiwe/avito-pr-service/internal/repository/pr"
	repo_team "github.com/4udiwe/avito-pr-service/internal/repository/team"
	repo_user "github.com/4udiwe/avito-pr-service/internal/repository/user"
	"github.com/4udiwe/avito-pr-service/internal/service/pr"
	"github.com/4udiwe/avito-pr-service/internal/service/team"
	"github.com/4udiwe/avito-pr-service/internal/service/user"
	"github.com/4udiwe/avito-pr-service/pkg/httpserver"
	"github.com/4udiwe/avito-pr-service/pkg/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

type App struct {
	cfg       *config.Config
	interrupt <-chan os.Signal

	// DB
	postgres *postgres.Postgres

	// Echo
	echoHandler *echo.Echo

	// Repositories
	userRepo *repo_user.Repository
	teamRepo *repo_team.Repository
	prRepo   *repo_pr.Repository

	// Handlers
	getPRsHandler         api.Handler
	getTeamHandler        api.Handler
	getTeamsHandler       api.Handler
	getUserReviewsHandler api.Handler

	postAssignUserToPRHandler   api.Handler
	postMergePRHandler          api.Handler
	postPRHandler               api.Handler
	postReassignReviewerHandler api.Handler
	postTeamHandler             api.Handler
	postIsUserActiveHandler     api.Handler

	// Services
	userService *user.Service
	teamService *team.Service
	prService   *pr.Service
}

func New(configPath string) *App {
	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("app - New - config.New: %v", err)
	}

	initLogger(cfg.Log.Level)

	return &App{
		cfg: cfg,
	}
}

func (app *App) Start() {
	// Postgres
	log.Info("Connecting to PostgreSQL...")

	postgres, err := postgres.New(app.cfg.Postgres.URL, postgres.ConnAttempts(5))

	if err != nil {
		log.Fatalf("app - Start - Postgres failed:%v", err)
	}
	app.postgres = postgres

	defer postgres.Close()

	// Migrations
	if err := database.RunMigrations(context.Background(), app.postgres.Pool); err != nil {
		log.Errorf("app - Start - Migrations failed: %v", err)
	}

	// App server
	log.Info("Starting app server...")
	httpServer := httpserver.New(app.EchoHandler(), httpserver.Port(app.cfg.HTTP.Port))
	httpServer.Start()
	log.Debugf("Server port: %s", app.cfg.HTTP.Port)

	defer func() {
		if err := httpServer.Shutdown(); err != nil {
			log.Errorf("HTTP server shutdown error: %v", err)
		}
	}()

	select {
	case s := <-app.interrupt:
		log.Infof("app - Start - signal: %v", s)
	case err := <-httpServer.Notify():
		log.Errorf("app - Start - server error: %v", err)
	}

	log.Info("Shutting down...")
}
