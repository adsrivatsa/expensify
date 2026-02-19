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
)

const categoriesCollection = "categories"

type mongoCategoryRepo struct {
	col *mongo.Collection
}

// NewCategoryRepository returns a MongoDB-backed CategoryRepository.
func NewCategoryRepository(db *mongo.Database) CategoryRepository {
	return &mongoCategoryRepo{col: db.Collection(categoriesCollection)}
}

func (r *mongoCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*models.Category, error) {
	cursor, err := r.col.Find(ctx, bson.M{"is_default": true})
	if err != nil {
		return nil, fmt.Errorf("category findDefaults: %w", err)
	}
	return decodeCategoryList(ctx, cursor)
}

func (r *mongoCategoryRepo) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Category, error) {
	cursor, err := r.col.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("category findByUserID: %w", err)
	}
	return decodeCategoryList(ctx, cursor)
}

func (r *mongoCategoryRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	var cat models.Category
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&cat)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("category findByID: %w", err)
	}
	return &cat, nil
}

func (r *mongoCategoryRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Category, error) {
	cursor, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("category findByIDs: %w", err)
	}
	return decodeCategoryList(ctx, cursor)
}

func (r *mongoCategoryRepo) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()

	if _, err := r.col.InsertOne(ctx, category); err != nil {
		return nil, fmt.Errorf("category create: %w", err)
	}
	return category, nil
}

// Delete removes a custom category only if it belongs to the given user.
func (r *mongoCategoryRepo) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": id, "user_id": userID, "is_default": false}
	result, err := r.col.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("category delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func decodeCategoryList(ctx context.Context, cursor *mongo.Cursor) ([]*models.Category, error) {
	defer cursor.Close(ctx)
	var categories []*models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, fmt.Errorf("category decode list: %w", err)
	}
	return categories, nil
}
