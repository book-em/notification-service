package test

import (
	"bookem-notification-service/internal"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkNotificationAsReadIntegration(t *testing.T) {
	guestUser := SetupGuestUser(t)
	hostUser := SetupHostUser(t)

	// Host creates a notification
	dto := internal.CreateNotificationDTO{
		ReceiverID: uint(guestUser.ID),
		Type:       "reservation_requested",
		Subject:    0,
		Object:     0,
	}
	resp, _ := CreateNotification(hostUser.JWT, dto)
	defer resp.Body.Close()
	notif := ResponseToNotification(resp)

	// Mark as read
	readResp, err := MarkNotificationAsRead(guestUser.JWT, notif.ID)
	require.NoError(t, err)
	defer readResp.Body.Close()

	require.Equal(t, http.StatusOK, readResp.StatusCode)

	// Verify notification is marked as read
	fetchResp, _ := GetUserNotifications(guestUser.JWT, 10, 0)
	defer fetchResp.Body.Close()
	notifs := ResponseToNotifications(fetchResp)
	var found bool
	for _, n := range notifs {
		if n.ID == notif.ID {
			found = n.IsRead
			break
		}
	}
	assert.True(t, found)
}
