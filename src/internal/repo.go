package internal

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	Create(ctx context.Context, notification *Notification) error
}

type repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{db}
}

func (r *repository) Create(ctx context.Context, notification *Notification) error {
	//todo implement this method
	_, err := r.db.Collection("notifications").InsertOne(ctx, notification)
	return err
}
