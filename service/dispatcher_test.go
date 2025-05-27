package service_test

import (
	"context"
	"testing"

	"github.com/iho/dispatcher/mock_client"
	"github.com/iho/dispatcher/model"
	"github.com/iho/dispatcher/service"
	"go.uber.org/mock/gomock"
)

func TestDispatcher(t *testing.T) {
	ctrl := gomock.NewController(t)

	config := service.Config{
		FetchURL: "http://example.com/api/users",
		PushURL:  "http://example.com/api/users",
		Suffix:   "@example.com",
	}

	users := model.PushUsers{
		model.PushUser{
			Email: "example@mail.com",
			Name:  "Example",
		},
		model.PushUser{
			Email: "example2@example.com",
			Name:  "Example2",
		},
		model.PushUser{
			Email: "example3@example.com",
			Name:  "Example3",
		},
		model.PushUser{
			Email: "example4@gmail.com",
			Name:  "Example4",
		},
	}

	fetchClient := mock_client.NewMockAPIFetchUsersClient(ctrl)
	fetchClient.EXPECT().FetchUsers(gomock.Any(), gomock.Any()).Return(
		users, nil).Times(1)

	pushClient := mock_client.NewMockAPIPushUsersClient(ctrl)

	pushClient.EXPECT().PushUsers(
		gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	dispatcher := service.NewDispatcher(
		fetchClient,
		pushClient,
		config,
	)

	filteredUsers, err := dispatcher.Dispatch(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(filteredUsers) != 2 {
		t.Fatalf("expected 3 users, got %d", len(filteredUsers))
	}

}
