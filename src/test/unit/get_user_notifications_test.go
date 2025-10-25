package test

import (
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetUserNotifications_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(2)).Return(&DefaultUser_Host, nil)
	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)

	expected := []internal.Notification{
		*DefaultNotification,
		{
			ID:         DefaultNotification.ID,
			ReceiverID: 2,
			Type:       internal.ReservationAccepted,
			Subject:    1,
			Object:     2,
			IsRead:     false,
		},
	}
	repo.On("FindByReceiverID", context.Background(), uint(2), 10, 0).Return(expected, nil)

	result, err := svc.GetUserNotifications(context.Background(), 2, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertCalled(t, "FindByReceiverID", context.Background(), uint(2), 10, 0)
	userClient.AssertCalled(t, "FindById", context.Background(), uint(2))
}

func Test_GetUserNotifications_UserNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.GetUserNotifications(context.Background(), 1, 10, 0)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, result)
	repo.AssertNotCalled(t, "FindByReceiverID")
}

func Test_GetUserNotifications_RepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)

	repo.On("FindByReceiverID", context.Background(), uint(1), 10, 0).Return(nil, errors.New("db error"))

	result, err := svc.GetUserNotifications(context.Background(), 1, 10, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertCalled(t, "FindByReceiverID", context.Background(), uint(1), 10, 0)
}
