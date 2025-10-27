package test

import (
	"bookem-notification-service/client/userclient"
	"bookem-notification-service/internal"
	"bookem-notification-service/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const URL_user = "http://user-service:8080/api/"
const URL_notification = "http://notification-service:8080/api/"

// ------------------------ Helpers ------------------------

func GenName(length int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[randInt(len(letterRunes))]
	}
	return string(b)
}

func randInt(max int) int {
	return int(time.Now().UnixNano() % int64(max))
}

// ------------------------ Users ------------------------

func RegisterUser(usernameOrEmail, password string, role util.UserRole) (*http.Response, error) {
	username := usernameOrEmail
	email := username + "@gmail.com"

	if strings.HasSuffix(usernameOrEmail, "@gmail.com") {
		username = strings.Split(usernameOrEmail, "@")[0]
		email = usernameOrEmail
	}

	dto := map[string]interface{}{
		"username": username,
		"password": password,
		"email":    email,
		"role":     string(role),
		"name":     GenName(6),
		"surname":  GenName(6),
		"address":  GenName(10),
	}

	jsonBytes, _ := json.Marshal(dto)
	return http.Post(URL_user+"register", "application/json", bytes.NewBuffer(jsonBytes))
}

func LoginUser(usernameOrEmail, password string) (*http.Response, error) {
	dto := map[string]string{
		"usernameOrEmail": usernameOrEmail,
		"password":        password,
	}
	jsonBytes, _ := json.Marshal(dto)
	return http.Post(URL_user+"login", "application/json", bytes.NewBuffer(jsonBytes))
}

func LoginUserJWT(usernameOrEmail, password string) string {
	resp, err := LoginUser(usernameOrEmail, password)
	if err != nil {
		panic(fmt.Sprintf("login failed: %v", err))
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var token userclient.JWTDTO
	if err := json.Unmarshal(bodyBytes, &token); err != nil {
		panic(fmt.Sprintf("failed to unmarshal jwt: %v", err))
	}
	return token.Jwt
}

// ------------------------ Notifications ------------------------

func CreateNotification(jwt string, dto internal.CreateNotificationDTO) (*http.Response, error) {
	jsonBytes, _ := json.Marshal(dto)
	req, _ := http.NewRequest(http.MethodPost, URL_notification+"notification", bytes.NewBuffer(jsonBytes))
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func GetUserNotifications(jwt string, limit, offset int) (*http.Response, error) {
	url := fmt.Sprintf("%snotifications?limit=%d&offset=%d", URL_notification, limit, offset)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func MarkNotificationAsRead(jwt, notificationID string) (*http.Response, error) {
	url := fmt.Sprintf("%snotifications/%s/read", URL_notification, notificationID)
	req, _ := http.NewRequest(http.MethodPut, url, nil)
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func GetUnreadNotificationCount(jwt string) (*http.Response, error) {
	url := URL_notification + "notifications/unread-count"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

// ------------------------ Notification Preferences ------------------------

func GetNotificationPreferences(jwt string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, URL_notification+"notification/preferences", nil)
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func UpdateNotificationPreferences(jwt string, enabledTypes map[internal.NotificationType]bool) (*http.Response, error) {
	payload := map[string]map[internal.NotificationType]bool{"enabledTypes": enabledTypes}
	jsonBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, URL_notification+"notification/preferences", bytes.NewBuffer(jsonBytes))
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

// ------------------------ Response Parsers ------------------------

func ResponseToNotification(resp *http.Response) internal.NotificationDTO {
	bodyBytes, _ := io.ReadAll(resp.Body)
	var obj internal.NotificationDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal notification: %v\nbody: %s", err, string(bodyBytes)))
	}
	return obj
}

func ResponseToNotifications(resp *http.Response) []internal.NotificationDTO {
	bodyBytes, _ := io.ReadAll(resp.Body)
	var obj []internal.NotificationDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal notifications: %v\nbody: %s", err, string(bodyBytes)))
	}
	return obj
}

func ResponseToUnreadCount(resp *http.Response) int64 {
	bodyBytes, _ := io.ReadAll(resp.Body)
	var obj struct {
		UnreadCount int64 `json:"unreadCount"`
	}
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal unread count: %v\nbody: %s", err, string(bodyBytes)))
	}
	return obj.UnreadCount
}

// ------------------------ Convenience Setup ------------------------

type UserInfo struct {
	ID       uint
	Username string
	Password string
	JWT      string
}

func SetupGuestUser(t *testing.T) UserInfo {
	username := "guest_" + GenName(5)
	password := "pass123"

	resp, err := RegisterUser(username, password, util.Guest)
	require.NoError(t, err)
	defer resp.Body.Close()

	var userResp struct {
		ID uint `json:"id"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &userResp)

	jwt := LoginUserJWT(username, password)
	return UserInfo{
		ID:       userResp.ID,
		Username: username,
		Password: password,
		JWT:      jwt,
	}
}

func SetupHostUser(t *testing.T) UserInfo {
	username := "host_" + GenName(5)
	password := "pass123"

	resp, err := RegisterUser(username, password, util.Host)
	require.NoError(t, err)
	defer resp.Body.Close()

	var userResp struct {
		ID uint `json:"id"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &userResp)

	jwt := LoginUserJWT(username, password)
	return UserInfo{
		ID:       userResp.ID,
		Username: username,
		Password: password,
		JWT:      jwt,
	}
}
