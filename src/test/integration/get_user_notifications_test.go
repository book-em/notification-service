package test

import (
	"bookem-notification-service/internal"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserNotificationsIntegration(t *testing.T) {
	guestUser := SetupGuestUser(t)
	hostUser := SetupHostUser(t)

	// Create 2 notifications for guest
	for i := 0; i < 2; i++ {
		dto := internal.CreateNotificationDTO{
			ReceiverID: uint(guestUser.ID),
			Type:       "reservation_requested",
			Subject:    0,
			Object:     0,
		}
		resp, _ := CreateNotification(hostUser.JWT, dto)
		defer resp.Body.Close()
	}

	// Fetch notifications
	resp, err := GetUserNotifications(guestUser.JWT, 10, 0)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	notifs := ResponseToNotifications(resp)
	assert.GreaterOrEqual(t, len(notifs), 2)
	for _, n := range notifs {
		assert.Equal(t, guestUser.ID, n.ReceiverID)
	}
}
