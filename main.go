package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/controller"
	"github.com/cubeface/cubeface/internal/model"
	"github.com/cubeface/cubeface/internal/router"
	"github.com/cubeface/cubeface/internal/service"
)

//go:embed all:web/dist
var webDist embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := model.InitDB(cfg); err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() {
		if err := model.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	sessions := service.NewSessionService(cfg)
	email := service.NewEmailService(cfg)
	verification := service.NewVerificationService(cfg, email)
	users := service.NewUserService(cfg, sessions, verification)
	oauth := service.NewOAuthService(cfg, users, sessions)

	userController := controller.NewUserController(cfg, users, sessions)
	verificationController := controller.NewVerificationController(verification)
	oauthController := controller.NewOAuthController(cfg, oauth, sessions)
	statusController := controller.NewStatusController(cfg, oauth)

	engine := router.New(router.Deps{
		Config:       cfg,
		Users:        userController,
		Verification: verificationController,
		OAuth:        oauthController,
		Status:       statusController,
		Sessions:     sessions,
		Static:       webDist,
	})

	if cfg.SessionSecret == config.DefaultDevSessionSecret {
		log.Printf("WARN using the development SESSION_SECRET; set SESSION_SECRET before deploying")
	}
	if !cfg.EmailVerificationEnabled {
		log.Printf("WARN email verification is disabled")
	}
	if providers := oauth.Providers(); len(providers) == 0 {
		log.Printf("oauth: no providers configured")
	} else {
		log.Printf("oauth: %d provider(s) configured", len(providers))
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("%s %s listening on http://localhost:%s%s", cfg.SystemName, cfg.Version, cfg.Port, cfg.BasePath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
	}
}
