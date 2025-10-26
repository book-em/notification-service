package internal

import (
	"context"
	"time"

	"bookem-notification-service/client/userclient"
	"bookem-notification-service/util"
)

type Service interface {
	CreateNotification(ctx context.Context, callerID uint, dto CreateNotificationDTO) (*Notification, error)
	GetUserNotifications(ctx context.Context, userID uint, limit int, offset int) ([]Notification, error)
	MarkNotificationAsRead(ctx context.Context, callerID uint, notificationID string) error
	GetUnreadNotificationCount(ctx context.Context, userID uint) (int64, error)
	GetUserNotificationPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID uint, enabledTypes map[NotificationType]bool) (*NotificationPreferences, error)
	CreateDefaultPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error)
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
func (s *service) CreateNotification(ctx context.Context, callerID uint, dto CreateNotificationDTO) (*Notification, error) {
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

func (s *service) GetUserNotifications(ctx context.Context, userID uint, limit int, offset int) ([]Notification, error) {
	util.TEL.Info("fetching notifications for user", nil, "user_id", userID, "limit", limit, "offset", offset)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	_, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Push(ctx, "find-user-notifications-in-db")
	defer util.TEL.Pop()
	return s.repo.FindByReceiverID(ctx, userID, limit, offset)
}

func (s *service) MarkNotificationAsRead(ctx context.Context, callerID uint, notificationID string) error {
	util.TEL.Info("mark notification as read", "caller_id", callerID, "notification_id", notificationID)

	util.TEL.Debug("check if user exists", nil, "id", callerID)
	_, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", callerID)
		return ErrUnauthenticated
	}

	util.TEL.Debug("check if notification exists", nil, "id", notificationID)
	notification, err := s.repo.FindByID(ctx, notificationID)
	if err != nil {
		util.TEL.Error("notification not found", err, "notification_id", notificationID)
		return ErrBadRequestCustom("notification does not exist")
	}

	util.TEL.Debug("check if user owns this notification", nil, "id", notificationID)
	if notification.ReceiverID != callerID {
		util.TEL.Error("user tried to mark notification not belonging to them", nil, "caller_id", callerID, "receiver_id", notification.ReceiverID)
		return ErrUnauthorized
	}

	if err := s.repo.MarkAsRead(ctx, notificationID); err != nil {
		util.TEL.Error("failed marking notification as read", err, "notification_id", notificationID)
		return err
	}

	util.TEL.Info("notification marked as read", "notification_id", notificationID)
	return nil
}

func (s *service) GetUnreadNotificationCount(ctx context.Context, userID uint) (int64, error) {
	util.TEL.Info("fetching unread notifications for user", nil, "user_id", userID)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	_, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return 0, ErrUnauthenticated
	}

	util.TEL.Push(ctx, "find-user-notifications-in-db")
	defer util.TEL.Pop()

	count, err := s.repo.CountUnreadNotifications(ctx, userID)
	if err != nil {
		util.TEL.Error("failed counting unread notifications", err, "user_id", userID)
		return 0, err
	}
	return count, nil
}

// ---------------------- Notification Preferences ----------------------

func (s *service) CreateDefaultPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error) {
	util.TEL.Info("ensuring default notification preferences", nil, "user_id", userID)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Debug("check if preferences already exist", nil, "user_id", userID)
	existing, err := s.repo.FindPreferencesByUserID(ctx, userID)
	if err != nil {
		util.TEL.Error("failed checking existing preferences", err, "user_id", userID)
		return nil, err
	}
	if existing != nil {
		util.TEL.Debug("preferences already exist, skipping default creation", nil, "user_id", userID)
		return existing, nil
	}

	defaultPrefs := make(map[NotificationType]bool)
	if user.Role == string(util.Host) {
		defaultPrefs[ReservationRequested] = true
		defaultPrefs[ReservationCancelled] = true
		defaultPrefs[HostReviewed] = true
		defaultPrefs[RoomReviewed] = true
	} else if user.Role == string(util.Guest) {
		defaultPrefs[ReservationAccepted] = true
		defaultPrefs[ReservationDeclined] = true
	} else {
		util.TEL.Error("unknown role for default preferences", nil, "role", user.Role)
		return nil, ErrUnauthorized
	}

	prefs := &NotificationPreferences{
		UserID:       userID,
		EnabledTypes: defaultPrefs,
	}

	if err := s.repo.SavePreferences(ctx, prefs); err != nil {
		util.TEL.Error("failed saving default preferences", err, "user_id", userID)
		return nil, err
	}

	util.TEL.Info("default notification preferences created", nil, "user_id", userID)
	return prefs, nil
}

func (s *service) GetUserNotificationPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error) {
	util.TEL.Info("fetching notification preferences for user", nil, "user_id", userID)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Debug("check if preferences exist", nil, "user_id", user.Id)
	prefs, err := s.repo.FindPreferencesByUserID(ctx, userID)
	if err != nil {
		util.TEL.Error("failed fetching preferences", err, "user_id", userID)
		return nil, err
	}

	if prefs == nil {
		util.TEL.Debug("no preferences found, creating default preferences", nil, "user_id", userID)
		prefs, err = s.CreateDefaultPreferences(ctx, userID)
		if err != nil {
			util.TEL.Error("failed creating default preferences", err, "user_id", userID)
			return nil, err
		}
	}

	return prefs, nil
}

func (s *service) UpdateNotificationPreferences(ctx context.Context, userID uint, enabledTypes map[NotificationType]bool) (*NotificationPreferences, error) {
	util.TEL.Info("updating notification preferences", nil, "user_id", userID)

	util.TEL.Debug("check if user exists", nil, "id", userID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), userID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", userID)
		return nil, ErrUnauthenticated
	}

	prefs, err := s.repo.FindPreferencesByUserID(ctx, user.Id)
	if err != nil {
		util.TEL.Error("failed fetching preferences", err, "user_id", userID)
		return nil, err
	}

	prefs.EnabledTypes = enabledTypes

	updatedPrefs, err := s.repo.UpdatePreferences(ctx, prefs)
	if err != nil {
		util.TEL.Error("failed updating preferences", err, "user_id", userID)
		return nil, err
	}

	util.TEL.Info("notification preferences updated", nil, "user_id", userID)
	return updatedPrefs, nil
}
