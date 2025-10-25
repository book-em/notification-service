package test

import (
	"bookem-notification-service/client/userclient"
	"bookem-notification-service/internal"
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTestNotificationService() (
	internal.Service,
	*MockNotificationRepo,
	*MockUserClient,
) {
	mockRepo := new(MockNotificationRepo)
	mockUserClient := new(MockUserClient)

	svc := internal.NewService(mockRepo, mockUserClient)
	return svc, mockRepo, mockUserClient
}

// ----------------------------------------------- Mock Notification repo

type MockNotificationRepo struct {
	mock.Mock
}

func (r *MockNotificationRepo) Save(ctx context.Context, notification *internal.Notification) (*internal.Notification, error) {
	args := r.Called(ctx, notification)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*internal.Notification), args.Error(1)
}

func (r *MockNotificationRepo) FindByReceiverID(ctx context.Context, receiverID uint, limit int, offset int) ([]internal.Notification, error) {
	args := r.Called(ctx, receiverID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]internal.Notification), args.Error(1)
}

func (r *MockNotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	args := r.Called(ctx, id)
	return args.Error(0)
}

func (r *MockNotificationRepo) FindByID(ctx context.Context, id string) (*internal.Notification, error) {
	args := r.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*internal.Notification), args.Error(1)
}

func (r *MockNotificationRepo) CountUnreadNotifications(ctx context.Context, receiverID uint) (int64, error) {
	args := r.Called(ctx, receiverID)
	return args.Get(0).(int64), args.Error(1)
}

// ----------------------------------------------- Mock user client

type MockUserClient struct {
	mock.Mock
}

func (r *MockUserClient) FindById(context context.Context, id uint) (*userclient.UserDTO, error) {
	args := r.Called(context, id)
	user, _ := args.Get(0).(*userclient.UserDTO)
	return user, args.Error(1)
}

// ----------------------------------------------- Mock data

var DefaultNotification = &internal.Notification{
	ID:         primitive.NewObjectID(),
	ReceiverID: 1,
	Type:       internal.ReservationRequested,
	Subject:    1,
	Object:     2,
	IsRead:     false,
}

var DefaultUser_Guest = userclient.UserDTO{
	Id:   1,
	Role: "guest",
}

var DefaultUser_Host = userclient.UserDTO{
	Id:   2,
	Role: "host",
}
