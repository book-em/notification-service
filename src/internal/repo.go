package internal

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	Save(ctx context.Context, notification *Notification) (*Notification, error)
	FindByReceiverID(ctx context.Context, receiverID uint, limit int, offset int) ([]Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	CountUnreadNotifications(ctx context.Context, receiverID uint) (int64, error)
	SavePreferences(ctx context.Context, prefs *NotificationPreferences) error
	UpdatePreferences(ctx context.Context, prefs *NotificationPreferences) (*NotificationPreferences, error)
	FindPreferencesByUserID(ctx context.Context, userID uint) (*NotificationPreferences, error)
}

type repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{db}
}

func (r *repository) Save(ctx context.Context, n *Notification) (*Notification, error) {
	coll := r.db.Collection("notifications")

	res, err := coll.InsertOne(ctx, n)
	if err != nil {
		return nil, err
	}

	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *repository) FindByReceiverID(ctx context.Context, receiverID uint, limit int, offset int) ([]Notification, error) {
	coll := r.db.Collection("notifications")

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := coll.Find(ctx, bson.M{"receiverId": receiverID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *repository) MarkAsRead(ctx context.Context, id string) error {
	coll := r.db.Collection("notifications")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = coll.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"isRead": true}},
	)
	return err
}

func (r *repository) FindByID(ctx context.Context, id string) (*Notification, error) {
	coll := r.db.Collection("notifications")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var notification Notification
	if err := coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&notification); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("notification not found: %w", err)
		}
		return nil, err
	}

	return &notification, nil
}

func (r *repository) CountUnreadNotifications(ctx context.Context, receiverID uint) (int64, error) {
	coll := r.db.Collection("notifications")

	filter := bson.M{
		"receiverId": receiverID,
		"isRead":     false,
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ---------------------- Notification Preferences ----------------------

func (r *repository) SavePreferences(ctx context.Context, prefs *NotificationPreferences) error {
	coll := r.db.Collection("notificationPreferences")

	enabledTypes := make(map[string]bool)
	for k, v := range prefs.EnabledTypes {
		enabledTypes[string(k)] = v
	}

	prefsDoc := bson.M{
		"userId": prefs.UserID,
		"types":  enabledTypes,
	}

	res, err := coll.InsertOne(ctx, prefsDoc)
	if err != nil {
		return err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		prefs.ID = oid
	}

	return nil
}
func (r *repository) UpdatePreferences(ctx context.Context, prefs *NotificationPreferences) (*NotificationPreferences, error) {
	coll := r.db.Collection("notificationPreferences")

	enabledTypes := make(map[string]bool, len(prefs.EnabledTypes))
	for k, v := range prefs.EnabledTypes {
		enabledTypes[string(k)] = v
	}

	update := bson.M{
		"$set": bson.M{
			"types": enabledTypes,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var result struct {
		ID     primitive.ObjectID `bson:"_id,omitempty"`
		UserID uint               `bson:"userId"`
		Types  map[string]bool    `bson:"types"`
	}

	err := coll.FindOneAndUpdate(ctx, bson.M{"userId": prefs.UserID}, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	enabledTyped := make(map[NotificationType]bool, len(result.Types))
	for k, v := range result.Types {
		enabledTyped[NotificationType(k)] = v
	}

	return &NotificationPreferences{
		ID:           result.ID,
		UserID:       result.UserID,
		EnabledTypes: enabledTyped,
	}, nil
}

func (r *repository) FindPreferencesByUserID(ctx context.Context, userID uint) (*NotificationPreferences, error) {
	coll := r.db.Collection("notificationPreferences")

	var result struct {
		ID     primitive.ObjectID `bson:"_id,omitempty"`
		UserID uint               `bson:"userId"`
		Types  map[string]bool    `bson:"types"`
	}

	err := coll.FindOne(ctx, bson.M{"userId": userID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	enabledTypes := make(map[NotificationType]bool, len(result.Types))
	for k, v := range result.Types {
		enabledTypes[NotificationType(k)] = v
	}

	return &NotificationPreferences{
		ID:           result.ID,
		UserID:       result.UserID,
		EnabledTypes: enabledTypes,
	}, nil
}
