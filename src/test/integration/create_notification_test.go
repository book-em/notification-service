package test

import (
	"bookem-notification-service/internal"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNotificationIntegration(t *testing.T) {
	hostUser := SetupHostUser(t)
	guestUser := SetupGuestUser(t)

	// Host creates a notification for guest
	dto := internal.CreateNotificationDTO{
		ReceiverID: uint(guestUser.ID),
		Type:       "reservation_requested",
		Subject:    0,
		Object:     0,
	}

	resp, err := CreateNotification(hostUser.JWT, dto)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	notif := ResponseToNotification(resp)
	assert.Equal(t, dto.Type, notif.Type)
	assert.Equal(t, dto.ReceiverID, notif.ReceiverID)
	assert.False(t, notif.IsRead)
}
