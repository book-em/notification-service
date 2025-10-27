package test

import (
	"bookem-notification-service/client/userclient"
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateDefaultPreferences_HostSuccess(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()
	userID := uint(1)

	userClient.On("FindById", mock.Anything, userID).Return(&DefaultUser_Host, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, userID).Return(nil, nil)
	repo.On("SavePreferences", mock.Anything, mock.AnythingOfType("*internal.NotificationPreferences")).Return(nil)

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
	assert.True(t, prefs.EnabledTypes[internal.ReservationRequested])
	assert.True(t, prefs.EnabledTypes[internal.ReservationCancelled])
	assert.True(t, prefs.EnabledTypes[internal.HostReviewed])
	assert.True(t, prefs.EnabledTypes[internal.RoomReviewed])
	repo.AssertCalled(t, "SavePreferences", mock.Anything, mock.Anything)
}

func Test_CreateDefaultPreferences_GuestSuccess(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()
	userID := uint(2)

	userClient.On("FindById", mock.Anything, userID).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, userID).Return(nil, nil)
	repo.On("SavePreferences", mock.Anything, mock.AnythingOfType("*internal.NotificationPreferences")).Return(nil)

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, userID, prefs.UserID)
	assert.True(t, prefs.EnabledTypes[internal.ReservationAccepted])
	assert.True(t, prefs.EnabledTypes[internal.ReservationDeclined])
	repo.AssertCalled(t, "SavePreferences", mock.Anything, mock.Anything)
}

func Test_CreateDefaultPreferences_AlreadyExists(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()
	userID := uint(3)

	existingPrefs := &internal.NotificationPreferences{
		UserID: userID,
		EnabledTypes: map[internal.NotificationType]bool{
			internal.ReservationAccepted: true,
		},
	}

	userClient.On("FindById", mock.Anything, userID).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, userID).Return(existingPrefs, nil)

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, existingPrefs, prefs)
	repo.AssertNotCalled(t, "SavePreferences", mock.Anything, mock.Anything)
}

func Test_CreateDefaultPreferences_UserNotFound(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()
	userID := uint(4)

	userClient.On("FindById", mock.Anything, userID).Return(nil, errors.New("not found"))

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.Nil(t, prefs)
	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
}

func Test_CreateDefaultPreferences_UnknownRole(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()
	userID := uint(5)

	userClient.On("FindById", mock.Anything, userID).Return(&userclient.UserDTO{
		Id:   userID,
		Role: "admin",
	}, nil)

	repo.On("FindPreferencesByUserID", mock.Anything, userID).Return(nil, nil)

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.Nil(t, prefs)
	assert.ErrorIs(t, err, internal.ErrUnauthorized)
}

func Test_CreateDefaultPreferences_SaveError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()
	userID := uint(6)

	userClient.On("FindById", mock.Anything, userID).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, userID).Return(nil, nil)
	repo.On("SavePreferences", mock.Anything, mock.AnythingOfType("*internal.NotificationPreferences")).Return(errors.New("db error"))

	prefs, err := svc.CreateDefaultPreferences(context.Background(), userID)

	assert.Nil(t, prefs)
	assert.EqualError(t, err, "db error")
}
