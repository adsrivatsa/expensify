package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const usersCollection = "users"

type mongoUserRepo struct {
	col *mongo.Collection
}

// NewUserRepository returns a MongoDB-backed UserRepository.
func NewUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepo{col: db.Collection(usersCollection)}
}

func (r *mongoUserRepo) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"google_id": googleID}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user findByGoogleID: %w", err)
	}
	return &user, nil
}

func (r *mongoUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user findByID: %w", err)
	}
	return &user, nil
}

// Upsert creates the user if they don't exist, or updates their profile fields.
func (r *mongoUserRepo) Upsert(ctx context.Context, user *models.User) (*models.User, error) {
	now := time.Now()
	user.UpdatedAt = now

	filter := bson.M{"google_id": user.GoogleID}
	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"name":       user.Name,
			"picture":    user.Picture,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"google_id":  user.GoogleID,
			"created_at": now,
		},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result models.User
	if err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		return nil, fmt.Errorf("user upsert: %w", err)
	}
	return &result, nil
}
