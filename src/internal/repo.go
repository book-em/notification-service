package internal

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	Save(ctx context.Context, notification *Notification) (*Notification, error)
	FindByReceiverID(ctx context.Context, receiverID string) ([]Notification, error)
	MarkAsRead(ctx context.Context, id string) error
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

func (r *repository) FindByReceiverID(ctx context.Context, receiverID string) ([]Notification, error) {
	coll := r.db.Collection("notifications")

	cursor, err := coll.Find(ctx, bson.M{"receiverId": receiverID})
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
