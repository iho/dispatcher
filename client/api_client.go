package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/iho/dispatcher/model"
)

type APIFetchUsersClient interface {
	FetchUsers(ctx context.Context, url string) (model.Users, error)
}

type APIFetchUsersClientJsonPlaceholderImpl struct {
	JsonPlaceHolderURL string
	Attempts           int
	httpClient         *http.Client
}

func NewAPIFetchUsersClientJsonPlaceholderImpl(url string, attempts int, httpClient *http.Client) *APIFetchUsersClientJsonPlaceholderImpl {
	return &APIFetchUsersClientJsonPlaceholderImpl{
		JsonPlaceHolderURL: url,
		Attempts:           attempts,
		httpClient:         httpClient,
	}
}

type APIPushUsersClient interface {
	PushUsers(ctx context.Context, users model.Users, url string) error
}

type APIPushUsersClientJsonPlaceholderImpl struct {
	PushDestinationURL string
	Attempts           int
	httpClient         *http.Client
}

func NewAPIPushUsersClientJsonPlaceholderImpl(url string, attempts int, httpClient *http.Client) *APIPushUsersClientJsonPlaceholderImpl {
	return &APIPushUsersClientJsonPlaceholderImpl{
		PushDestinationURL: url,
		Attempts:           attempts,
		httpClient:         httpClient,
	}
}
func (c *APIFetchUsersClientJsonPlaceholderImpl) FetchUsers(ctx context.Context, url string) (model.Users, error) {

	var users model.Users
	err := Retry(c.Attempts, time.Second, func() error {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		if resp.Body == nil {
			return fmt.Errorf("response body is nil")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		if len(body) == 0 {
			return fmt.Errorf("response body is empty")
		}
		slog.Info("Fetched users from API", "url", url, "length", len(body))
		users, err = model.UnmarshalUsers(body)
		if err != nil {
			return fmt.Errorf("failed to unmarshal users: %w", err)
		}
		return nil
	})
	return users, err
}
func (c *APIPushUsersClientJsonPlaceholderImpl) PushUsers(ctx context.Context, users model.Users, url string) error {
	jsonData, err := users.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	err = Retry(c.Attempts, time.Second, func() error {
		req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		if resp.Body == nil {
			return fmt.Errorf("response body is nil")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		if len(body) == 0 {
			return fmt.Errorf("response body is empty")
		}
		slog.Info("Pushed users to API", "url", url, "length", len(body))
		return nil
	})
	return err
}

func Retry(attempts int, sleep time.Duration, fn func() error) error {
	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		slog.Error("Retrying after error", "attempt", i+1, "error", err)
		time.Sleep(sleep)
		sleep *= 2 // exponential backoff
	}
	return fmt.Errorf("after %d attempts, last error: %w", attempts, fn())
}
