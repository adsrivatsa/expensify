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

const sessionsCollection = "sessions"

type mongoSessionRepo struct {
	col *mongo.Collection
}

// NewSessionRepository returns a MongoDB-backed SessionRepository.
func NewSessionRepository(db *mongo.Database) SessionRepository {
	return &mongoSessionRepo{col: db.Collection(sessionsCollection)}
}

func (r *mongoSessionRepo) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	session.ID = primitive.NewObjectID()
	session.CreatedAt = time.Now()

	if _, err := r.col.InsertOne(ctx, session); err != nil {
		return nil, fmt.Errorf("session create: %w", err)
	}
	return session, nil
}

func (r *mongoSessionRepo) FindByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	err := r.col.FindOne(ctx, bson.M{"token": token}).Decode(&session)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session findByToken: %w", err)
	}
	return &session, nil
}

func (r *mongoSessionRepo) Delete(ctx context.Context, token string) error {
	if _, err := r.col.DeleteOne(ctx, bson.M{"token": token}); err != nil {
		return fmt.Errorf("session delete: %w", err)
	}
	return nil
}

func (r *mongoSessionRepo) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lt": time.Now()}}
	if _, err := r.col.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("session deleteExpired: %w", err)
	}
	return nil
}

// EnsureIndexes creates the TTL index on sessions so MongoDB auto-expires them.
func EnsureSessionIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection(sessionsCollection)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
}
