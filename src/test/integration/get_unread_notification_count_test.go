package test

import (
	"bookem-notification-service/internal"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUnreadNotificationCountIntegration(t *testing.T) {
	guestUser := SetupGuestUser(t)
	hostUser := SetupHostUser(t)

	// Create 3 notifications
	for i := 0; i < 3; i++ {
		dto := internal.CreateNotificationDTO{
			ReceiverID: uint(guestUser.ID),
			Type:       "reservation_declined",
			Subject:    0,
			Object:     0,
		}
		resp, _ := CreateNotification(hostUser.JWT, dto)
		defer resp.Body.Close()
	}

	// Get unread count
	resp, err := GetUnreadNotificationCount(guestUser.JWT)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	count := ResponseToUnreadCount(resp)
	assert.GreaterOrEqual(t, count, int64(3))
}
