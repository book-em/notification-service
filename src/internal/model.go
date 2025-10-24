package internal

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	ReservationRequested NotificationType = "reservation_requested"
	ReservationCancelled NotificationType = "reservation_cancelled"
	HostReviewed         NotificationType = "host_reviewed"
	RoomReviewed         NotificationType = "room_reviewed"
	ReservationAccepted  NotificationType = "reservation_accepted"
	ReservationDeclined  NotificationType = "reservation_declined"
)

type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ReceiverID  string             `bson:"receiverId" json:"receiverId"`
	Type        NotificationType   `bson:"type" json:"type"`
	Subject     string             `bson:"subject" json:"subject"`
	Object      string             `bson:"object,omitempty" json:"object"`
	StarsNumber int                `bson:"starsNumber,omitempty" json:"starsNumber,omitempty"`
	IsRead      bool               `bson:"isRead" json:"isRead"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

type NotificationPreferences struct {
	ID           primitive.ObjectID        `bson:"_id,omitempty" json:"id"`
	UserID       string                    `bson:"userId" json:"userId"`
	EnabledTypes map[NotificationType]bool `bson:"types" json:"types"`
}
