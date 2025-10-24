package internal

import (
	"context"

	"bookem-notification-service/client/userclient"
)

type Service interface {
	Create(ctx context.Context, callerID uint, dto NotificationDTO) (*Notification, error)
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

func (s *service) Create(ctx context.Context, callerID uint, dto NotificationDTO) (*Notification, error) {
	notification := &Notification{}
	return notification, s.repo.Create(ctx, notification)
}
