package test

import (
	"bookem-notification-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_GetUserNotificationPreferences_SuccessExistingPrefs(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)

	existingPrefs := &internal.NotificationPreferences{
		ID:     primitive.NewObjectID(),
		UserID: 1,
		EnabledTypes: map[internal.NotificationType]bool{
			internal.ReservationRequested: true,
			internal.HostReviewed:         false,
		},
	}
	repo.On("FindPreferencesByUserID", context.Background(), uint(1)).Return(existingPrefs, nil)

	prefs, err := svc.GetUserNotificationPreferences(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, existingPrefs, prefs)
	repo.AssertCalled(t, "FindPreferencesByUserID", context.Background(), uint(1))
	userClient.AssertCalled(t, "FindById", context.Background(), uint(1))
}

func Test_GetUserNotificationPreferences_UserNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	prefs, err := svc.GetUserNotificationPreferences(context.Background(), 1)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, prefs)
	repo.AssertNotCalled(t, "FindPreferencesByUserID")
}

func Test_GetUserNotificationPreferences_RepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", context.Background(), uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", context.Background(), uint(1)).Return(nil, errors.New("db error"))

	prefs, err := svc.GetUserNotificationPreferences(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, prefs)
	repo.AssertCalled(t, "FindPreferencesByUserID", context.Background(), uint(1))
}

func Test_GetUserNotificationPreferences_CreateDefaultPrefs(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	userClient.On("FindById", mock.Anything, uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, uint(1)).Return(nil, nil)
	repo.On("SavePreferences", mock.Anything, mock.AnythingOfType("*internal.NotificationPreferences")).Return(nil)

	prefs, err := svc.GetUserNotificationPreferences(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, uint(1), prefs.UserID)
	assert.True(t, prefs.EnabledTypes[internal.ReservationAccepted])
	assert.True(t, prefs.EnabledTypes[internal.ReservationDeclined])

	repo.AssertCalled(t, "FindPreferencesByUserID", mock.Anything, uint(1))
	repo.AssertCalled(t, "SavePreferences", mock.Anything, mock.AnythingOfType("*internal.NotificationPreferences"))
	userClient.AssertCalled(t, "FindById", mock.Anything, uint(1))
}
