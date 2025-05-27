package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/iho/dispatcher/client"
	"github.com/iho/dispatcher/service"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file", err)
	}

	var cfg service.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal("Error processing env vars into config struct:", err)
	}

	slog.Info("Configuration loaded", "config", cfg)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	dispatcher := service.NewDispatcher(
		client.NewAPIFetchUsersClientJsonPlaceholderImpl(cfg.FetchURL, cfg.Attempts, httpClient),
		client.NewAPIPushUsersClientJsonPlaceholderImpl(cfg.PushURL, cfg.Attempts, httpClient),
		cfg,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // 10 seconds
	filteredUsers, err := dispatcher.Dispatch(ctx)
	if err != nil {
		log.Fatal("Error dispatching:", err)
	}
	slog.Info("Filtered users", "users", filteredUsers)

	slog.Info("Done")
}
