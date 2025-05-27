package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iho/dispatcher/client"
	"github.com/iho/dispatcher/model"
)

type Dispatcher interface {
	Dispatch(ctx context.Context) (model.PushUsers, error)
}

// Config holds the configuration for the Dispatcher.
type Config struct {
	FetchURL string `envconfig:"FETCH_URL"`
	PushURL  string `envconfig:"PUSH_URL"`
	Suffix   string `envconfig:"SUFFIX"`
	Interval int    `envconfig:"INTERVAL"`
	Attempts int    `envconfig:"ATTEMPTS"`
}

type DispatcherImpl struct {
	FetchClient      client.APIFetchUsersClient
	PushClient       client.APIPushUsersClient
	DispatcherConfig Config
}

func (d *DispatcherImpl) Dispatch(ctx context.Context) (model.PushUsers, error) {
	users, err := d.FetchClient.FetchUsers(ctx, d.DispatcherConfig.FetchURL)
	if err != nil {
		slog.Error("Failed to fetch users", "error", err)
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	filteredUsers := FilterUsers(users, d.DispatcherConfig.Suffix)

	slog.Info("Filtered users", "count", len(filteredUsers))

	// Push users to the API
	err = d.PushClient.PushUsers(ctx, filteredUsers, d.DispatcherConfig.PushURL)
	if err != nil {
		slog.Error("Failed to push users", "error", err)
		return nil, fmt.Errorf("failed to push users: %w", err)
	}

	slog.Info("Pushed users", "count", len(filteredUsers))
	return filteredUsers, nil
}

func FilterUsers(users []model.User, suffix string) model.PushUsers {
	var pushUsers model.PushUsers
	for _, user := range users {
		if strings.HasSuffix(user.Email, suffix) {
			pushUsers = append(pushUsers, model.PushUser{
				Email: user.Email,
				Name:  user.Name,
			})
		} else {
			slog.Info("User does not match suffix", "email", user.Email)
		}
	}
	return pushUsers
}

// NewDispatcher creates a new Dispatcher instance with the provided clients and configuration.
func NewDispatcher(fetchClient client.APIFetchUsersClient, pushClient client.APIPushUsersClient, config Config) Dispatcher {
	return &DispatcherImpl{
		FetchClient:      fetchClient,
		PushClient:       pushClient,
		DispatcherConfig: config,
	}
}
