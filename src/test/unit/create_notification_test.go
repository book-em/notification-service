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

func Test_CreateNotification_Success(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{
		ReceiverID:  2,
		Type:        internal.ReservationRequested,
		Subject:     100,
		Object:      200,
		StarsNumber: 0,
	}

	guestUser := userclient.UserDTO{Id: callerID, Role: "guest"}
	hostUser := userclient.UserDTO{Id: dto.ReceiverID, Role: "host"}

	userClient.On("FindById", mock.Anything, callerID).Return(&guestUser, nil)
	userClient.On("FindById", mock.Anything, dto.ReceiverID).Return(&hostUser, nil)

	receiverPrefs := &internal.NotificationPreferences{
		UserID: dto.ReceiverID,
		EnabledTypes: map[internal.NotificationType]bool{
			internal.ReservationRequested: true,
		},
	}
	repo.On("FindPreferencesByUserID", mock.Anything, dto.ReceiverID).Return(receiverPrefs, nil)

	expected := &internal.Notification{
		ReceiverID:  dto.ReceiverID,
		Type:        dto.Type,
		Subject:     dto.Subject,
		Object:      dto.Object,
		StarsNumber: dto.StarsNumber,
		IsRead:      false,
	}
	repo.On("Save", mock.Anything, mock.AnythingOfType("*internal.Notification")).Return(expected, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertCalled(t, "Save", mock.Anything, mock.AnythingOfType("*internal.Notification"))
	userClient.AssertCalled(t, "FindById", mock.Anything, callerID)
	userClient.AssertCalled(t, "FindById", mock.Anything, dto.ReceiverID)
	repo.AssertCalled(t, "FindPreferencesByUserID", mock.Anything, dto.ReceiverID)
}
func Test_CreateNotification_DisabledPreference(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{
		ReceiverID:  2,
		Type:        internal.ReservationRequested,
		Subject:     100,
		Object:      200,
		StarsNumber: 0,
	}

	guestUser := userclient.UserDTO{Id: callerID, Role: "guest"}
	hostUser := userclient.UserDTO{Id: dto.ReceiverID, Role: "host"}

	userClient.On("FindById", mock.Anything, callerID).Return(&guestUser, nil)
	userClient.On("FindById", mock.Anything, dto.ReceiverID).Return(&hostUser, nil)

	receiverPrefs := &internal.NotificationPreferences{
		UserID: dto.ReceiverID,
		EnabledTypes: map[internal.NotificationType]bool{
			internal.ReservationCancelled: true,
		},
	}
	repo.On("FindPreferencesByUserID", mock.Anything, dto.ReceiverID).Return(receiverPrefs, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "notification type is disabled")
}

func Test_CreateNotification_PreferencesRepoError(t *testing.T) {
	svc, repo, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{
		ReceiverID:  2,
		Type:        internal.ReservationRequested,
		Subject:     100,
		Object:      200,
		StarsNumber: 0,
	}

	guestUser := userclient.UserDTO{Id: callerID, Role: "guest"}
	hostUser := userclient.UserDTO{Id: dto.ReceiverID, Role: "host"}

	userClient.On("FindById", mock.Anything, callerID).Return(&guestUser, nil)
	userClient.On("FindById", mock.Anything, dto.ReceiverID).Return(&hostUser, nil)

	repo.On("FindPreferencesByUserID", mock.Anything, dto.ReceiverID).
		Return(nil, errors.New("db error"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
}

func Test_CreateNotification_UserNotFound(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(nil, errors.New("not found"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, result)
}

func Test_CreateNotification_InvalidRole(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

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
	dto := internal.CreateNotificationDTO{ReceiverID: 0, Type: internal.ReservationRequested}

	userClient.On("FindById", context.Background(), callerID).Return(&DefaultUser_Guest, nil)

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receiver ID cannot be empty")
	assert.Nil(t, result)
}

func Test_CreateNotification_ReceiverNotFound(t *testing.T) {
	svc, _, userClient := CreateTestNotificationService()

	callerID := uint(1)
	dto := internal.CreateNotificationDTO{ReceiverID: 2, Type: internal.ReservationRequested}

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
	dto := internal.CreateNotificationDTO{ReceiverID: 2, Type: ""}

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
	dto := internal.CreateNotificationDTO{
		ReceiverID: 2,
		Type:       internal.ReservationRequested,
	}

	userClient.On("FindById", mock.Anything, callerID).Return(&DefaultUser_Guest, nil)
	userClient.On("FindById", mock.Anything, dto.ReceiverID).Return(&DefaultUser_Host, nil)

	repo.On("FindPreferencesByUserID", mock.Anything, dto.ReceiverID).
		Return(&internal.NotificationPreferences{
			UserID: dto.ReceiverID,
			EnabledTypes: map[internal.NotificationType]bool{
				internal.ReservationRequested: true,
			},
		}, nil)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*internal.Notification")).
		Return(nil, errors.New("db error"))

	result, err := svc.CreateNotification(context.Background(), callerID, dto)

	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
	assert.Nil(t, result)

	repo.AssertCalled(t, "FindPreferencesByUserID", mock.Anything, dto.ReceiverID)
	repo.AssertCalled(t, "Save", mock.Anything, mock.AnythingOfType("*internal.Notification"))
}
