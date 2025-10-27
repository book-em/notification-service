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

func Test_UpdateNotificationPreferences_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	enabledTypes := map[internal.NotificationType]bool{
		internal.ReservationRequested: true,
		internal.ReservationCancelled: false,
	}

	userClient.On("FindById", mock.Anything, uint(1)).Return(&DefaultUser_Guest, nil)

	existingPrefs := &internal.NotificationPreferences{
		ID:           primitive.NewObjectID(),
		UserID:       1,
		EnabledTypes: map[internal.NotificationType]bool{},
	}
	repo.On("FindPreferencesByUserID", mock.Anything, uint(1)).Return(existingPrefs, nil)

	updatedPrefs := &internal.NotificationPreferences{
		ID:           existingPrefs.ID,
		UserID:       1,
		EnabledTypes: enabledTypes,
	}
	repo.On("UpdatePreferences", mock.Anything, existingPrefs).Return(updatedPrefs, nil)

	result, err := svc.UpdateNotificationPreferences(context.Background(), 1, enabledTypes)

	assert.NoError(t, err)
	assert.Equal(t, updatedPrefs, result)
	userClient.AssertCalled(t, "FindById", mock.Anything, uint(1))
	repo.AssertCalled(t, "FindPreferencesByUserID", mock.Anything, uint(1))
	repo.AssertCalled(t, "UpdatePreferences", mock.Anything, existingPrefs)
}

func Test_UpdateNotificationPreferences_UserNotFound(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	enabledTypes := map[internal.NotificationType]bool{
		internal.ReservationRequested: true,
	}

	userClient.On("FindById", mock.Anything, uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.UpdateNotificationPreferences(context.Background(), 1, enabledTypes)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, result)
	repo.AssertNotCalled(t, "FindPreferencesByUserID")
	repo.AssertNotCalled(t, "UpdatePreferences")
}

func Test_UpdateNotificationPreferences_FindPrefsError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	enabledTypes := map[internal.NotificationType]bool{
		internal.ReservationRequested: true,
	}

	userClient.On("FindById", mock.Anything, uint(1)).Return(&DefaultUser_Guest, nil)
	repo.On("FindPreferencesByUserID", mock.Anything, uint(1)).Return(nil, errors.New("db error"))

	result, err := svc.UpdateNotificationPreferences(context.Background(), 1, enabledTypes)

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertCalled(t, "FindPreferencesByUserID", mock.Anything, uint(1))
	repo.AssertNotCalled(t, "UpdatePreferences")
}

func Test_UpdateNotificationPreferences_UpdatePrefsError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	enabledTypes := map[internal.NotificationType]bool{
		internal.ReservationRequested: true,
	}

	userClient.On("FindById", mock.Anything, uint(1)).Return(&DefaultUser_Guest, nil)

	existingPrefs := &internal.NotificationPreferences{
		ID:           primitive.NewObjectID(),
		UserID:       1,
		EnabledTypes: map[internal.NotificationType]bool{},
	}
	repo.On("FindPreferencesByUserID", mock.Anything, uint(1)).Return(existingPrefs, nil)

	repo.On("UpdatePreferences", mock.Anything, existingPrefs).Return(nil, errors.New("update failed"))

	result, err := svc.UpdateNotificationPreferences(context.Background(), 1, enabledTypes)

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertCalled(t, "UpdatePreferences", mock.Anything, existingPrefs)
}
