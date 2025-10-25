package internal

import "time"

type CreateNotificationDTO struct {
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

func NewNotificationDTO(n *Notification) NotificationDTO {
	return NotificationDTO{
		ID:          n.ID.Hex(),
		ReceiverID:  n.ReceiverID,
		Type:        n.Type,
		Subject:     n.Subject,
		Object:      n.Object,
		StarsNumber: n.StarsNumber,
		IsRead:      n.IsRead,
		CreatedAt:   n.CreatedAt,
	}
}

type NotificationPreferencesDTO struct {
	UserID       uint                      `json:"userId"`
	EnabledTypes map[NotificationType]bool `json:"enabledTypes"`
}
