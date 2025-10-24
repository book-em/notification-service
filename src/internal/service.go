package internal

import (
	"context"
	"time"

	"bookem-notification-service/client/userclient"
	"bookem-notification-service/util"
)

type Service interface {
	CreateNotification(ctx context.Context, callerID uint, dto NewNotificationDTO) (*Notification, error)
	GetUserNotifications(ctx context.Context, userID uint, limit int) ([]Notification, error)
}

type service struct {
	repo       Repository
	userClient userclient.UserClient
}

func NewService(
	repo Repository,
	userClient userclient.UserClient) Service {
	return &service{repo, userClient}
}

// CreateNotification - stores a new notification in MongoDB
func (s *service) CreateNotification(ctx context.Context, callerID uint, dto NewNotificationDTO) (*Notification, error) {
	util.TEL.Info("user initiates creating a notification", nil, "caller_id", callerID)

	util.TEL.Debug("check if user exists", nil, "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", callerID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Debug("check if user is a guest or host", nil, "id", callerID)
	if user.Role != string(util.Guest) {
		if user.Role != string(util.Host) {
			util.TEL.Error("user has a bad role", nil, "role", user.Role)
			return nil, ErrUnauthorized
		}
	}

	util.TEL.Info("creating new notification", nil, "receiver_id", dto.ReceiverID, "type", dto.Type)

	if dto.ReceiverID == 0 {
		util.TEL.Error("missing receiver ID", nil)
		return nil, ErrBadRequestCustom("receiver ID cannot be empty")
	}

	util.TEL.Debug("check if receiver exists", nil, "id", dto.ReceiverID)
	receiver, err := s.userClient.FindById(util.TEL.Ctx(), dto.ReceiverID)
	if err != nil {
		util.TEL.Error("receiver does not exist", err, "id", dto.ReceiverID)
		return nil, ErrBadRequestCustom("receiver does not exist")
	}

	if dto.Type == "" {
		util.TEL.Error("missing notification type", nil)
		return nil, ErrBadRequestCustom("notification type cannot be empty")
	}

	notification := &Notification{
		ReceiverID:  receiver.Id,
		Type:        dto.Type,
		Subject:     dto.Subject,
		Object:      dto.Object,
		StarsNumber: dto.StarsNumber,
		IsRead:      false,
		CreatedAt:   time.Now(),
	}

	saved, err := s.repo.Save(ctx, notification)
	if err != nil {
		util.TEL.Error("failed to save notification", err)
		return nil, err
	}

	util.TEL.Info("notification successfully created", "notification_id", saved.ID)
	return saved, nil
}

func (s *service) GetUserNotifications(ctx context.Context, userID uint, limit int) ([]Notification, error) {
	util.TEL.Info("fetching notifications for user", nil, "user_id", userID, "limit", limit)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	_, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Push(ctx, "find-user-notifications-in-db")
	defer util.TEL.Pop()
	return s.repo.FindByReceiverID(ctx, userID, limit)
}
