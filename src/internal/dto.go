package internal

import "time"

type NewNotificationDTO struct {
	ReceiverID  uint             `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     uint             `json:"subject"`
	Object      uint             `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
}

type NotificationDTO struct {
	ID          string           `json:"id"`
	ReceiverID  uint             `json:"receiverId"`
	Type        NotificationType `json:"type"`
	Subject     uint             `json:"subject"`
	Object      uint             `json:"object"`
	StarsNumber int              `json:"starsNumber,omitempty"`
	IsRead      bool             `json:"isRead"`
	CreatedAt   time.Time        `json:"createdAt"`
}

type NotificationPreferencesDTO struct {
	UserID       uint                      `json:"userId"`
	EnabledTypes map[NotificationType]bool `json:"enabledTypes"`
}
