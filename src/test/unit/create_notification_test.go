package test

import (
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateNotification_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{
		ReceiverID:  2,
		Type:        internal.ReservationRequested,
		Subject:     100,
		Object:      200,
		StarsNumber: 0,
	}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)
	userClient.On("FindById", context.Background(), dto.ReceiverID).Return(&DefaultUser_Host, nil)

	expected := &internal.Notification{
		ReceiverID:  dto.ReceiverID,
		Type:        dto.Type,
		Subject:     dto.Subject,
		Object:      dto.Object,
		StarsNumber: dto.StarsNumber,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}
	repo.On("Save", context.Background(), mock.AnythingOfType("*internal.Notification")).Return(expected, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertCalled(t, "Save", context.Background(), mock.AnythingOfType("*internal.Notification"))
	userClient.AssertCalled(t, "FindById", context.Background(), callerID)
	userClient.AssertCalled(t, "FindById", context.Background(), dto.ReceiverID)
}

func Test_CreateNotification_UserNotFound(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(nil, errors.New("not found"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, result)
}

func Test_CreateNotification_InvalidRole(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

	invalidUser := DefaultUser_Guest
	invalidUser.Role = "admin"
	userClient.On("FindById", context.Background(), callerID).Return(&invalidUser, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}

func Test_CreateNotification_MissingReceiverID(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 0, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver ID cannot be empty")
	assert.Nil(t, result)
}

func Test_CreateNotification_ReceiverNotFound(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)
	userClient.On("FindById", context.Background(), dto.ReceiverID).Return(nil, errors.New("not found"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver does not exist")
	assert.Nil(t, result)
}

func Test_CreateNotification_MissingType(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 2, Type: ""}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)
	userClient.On("FindById", context.Background(), dto.ReceiverID).Return(&DefaultUser_Host, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification type cannot be empty")
	assert.Nil(t, result)
}

func Test_CreateNotification_RepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.NewNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)
	userClient.On("FindById", context.Background(), dto.ReceiverID).Return(&DefaultUser_Host, nil)

	repo.On("Save", context.Background(), mock.AnythingOfType("*internal.Notification")).Return(nil, errors.New("db error"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
	assert.Nil(t, result)
}
