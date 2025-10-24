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
