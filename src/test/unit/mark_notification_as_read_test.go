package test

import (
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MarkNotificationAsRead_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	notificationID := DefaultNotification.ID.Hex()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindByID", context.Background(), notificationID).Return(DefaultNotification, nil)
	repo.On("MarkAsRead", context.Background(), notificationID).Return(nil)

	err := svc.MarkNotificationAsRead(context.Background(), 1, notificationID)

	assert.NoError(t, err)
	repo.AssertCalled(t, "FindByID", context.Background(), notificationID)
	repo.AssertCalled(t, "MarkAsRead", context.Background(), notificationID)
	userClient.AssertCalled(t, "FindById", context.Background(), uint(1))
}

func Test_MarkNotificationAsRead_UserNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	err := svc.MarkNotificationAsRead(context.Background(), 1, "some-id")

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	repo.AssertNotCalled(t, "FindByID")
	repo.AssertNotCalled(t, "MarkAsRead")
}

func Test_MarkNotificationAsRead_NotificationNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	notificationID := "nonexistent-id"

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindByID", context.Background(), notificationID).Return(nil, errors.New("not found"))

	err := svc.MarkNotificationAsRead(context.Background(), 1, notificationID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification does not exist")
	repo.AssertCalled(t, "FindByID", context.Background(), notificationID)
	repo.AssertNotCalled(t, "MarkAsRead")
}

func Test_MarkNotificationAsRead_UnauthorizedUser(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	notificationID := DefaultNotification.ID.Hex()

	userClient.On("FindById", context.Background(), uint(2)).Return(&DefaultUser_Host, nil)

	foreignNotification := *DefaultNotification
	foreignNotification.ReceiverID = 99
	repo.On("FindByID", context.Background(), notificationID).Return(&foreignNotification, nil)

	err := svc.MarkNotificationAsRead(context.Background(), 2, notificationID)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	repo.AssertCalled(t, "FindByID", context.Background(), notificationID)
	repo.AssertNotCalled(t, "MarkAsRead")
}

func Test_MarkNotificationAsRead_RepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	notificationID := DefaultNotification.ID.Hex()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindByID", context.Background(), notificationID).Return(DefaultNotification, nil)
	repo.On("MarkAsRead", context.Background(), notificationID).Return(errors.New("db error"))

	err := svc.MarkNotificationAsRead(context.Background(), 1, notificationID)

	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
	repo.AssertCalled(t, "MarkAsRead", context.Background(), notificationID)
}
