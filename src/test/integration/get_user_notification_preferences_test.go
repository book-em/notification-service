package test

import (
	"bookem-notification-service/internal"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserNotificationPreferencesIntegration(t *testing.T) {
	guestUser := SetupGuestUser(t)

	resp, err := GetNotificationPreferences(guestUser.JWT)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)
	var prefs internal.NotificationPreferencesDTO
	err = json.Unmarshal(bodyBytes, &prefs)
	require.NoError(t, err)

	assert.Equal(t, guestUser.ID, prefs.UserID)
	assert.NotEmpty(t, prefs.EnabledTypes)
}
