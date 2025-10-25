package test

import (
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetUnreadNotificationCount_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)

	repo.On("CountUnreadNotifications", context.Background(), uint(1)).Return(int64(5), nil)

	count, err := svc.GetUnreadNotificationCount(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
	repo.AssertCalled(t, "CountUnreadNotifications", context.Background(), uint(1))
	userClient.AssertCalled(t, "FindById", context.Background(), uint(1))
}

func Test_GetUnreadNotificationCount_UserNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	count, err := svc.GetUnreadNotificationCount(context.Background(), 1)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Equal(t, int64(0), count)
	repo.AssertNotCalled(t, "CountUnreadNotifications")
}

func Test_GetUnreadNotificationCount_RepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)

	repo.On("CountUnreadNotifications", context.Background(), uint(1)).Return(int64(0), errors.New("db error"))

	count, err := svc.GetUnreadNotificationCount(context.Background(), 1)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	repo.AssertCalled(t, "CountUnreadNotifications", context.Background(), uint(1))
}
