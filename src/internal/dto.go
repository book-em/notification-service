package internal

import "time"

type NewNotificationDTO struct {
	ReceiverID  string           `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     string           `json:"subject"`
	Object      string           `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
}

type NotificationDTO struct {
	ID          string           `json:"id"`
	ReceiverID  string           `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     string           `json:"subject"`
	Object      string           `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
	IsRead      bool             `json:"isRead"`
	CreatedAt   time.Time        `json:"createdAt"`
}

type NotificationPreferencesDTO struct {
	UserID       string                    `json:"userId"`
	EnabledTypes map[NotificationType]bool `json:"enabledTypes"`
}
