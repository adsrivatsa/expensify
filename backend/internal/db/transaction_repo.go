package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const transactionsCollection = "transactions"

// ErrNotFound is returned when a document is not found or the caller has no access to it.
var ErrNotFound = errors.New("not found")

type mongoTransactionRepo struct {
	col *mongo.Collection
}

// NewTransactionRepository returns a MongoDB-backed TransactionRepository.
func NewTransactionRepository(db *mongo.Database) TransactionRepository {
	return &mongoTransactionRepo{col: db.Collection(transactionsCollection)}
}

func (r *mongoTransactionRepo) Create(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	tx.ID = primitive.NewObjectID()
	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	if _, err := r.col.InsertOne(ctx, tx); err != nil {
		return nil, fmt.Errorf("transaction create: %w", err)
	}
	return tx, nil
}

func (r *mongoTransactionRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&tx)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("transaction findByID: %w", err)
	}
	return &tx, nil
}

// FindByUserID returns a paginated, date-descending list of transactions for a user.
func (r *mongoTransactionRepo) FindByUserID(
	ctx context.Context,
	userID primitive.ObjectID,
	page, pageSize int,
) ([]*models.Transaction, int64, error) {
	filter := bson.M{"user_id": userID}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction count: %w", err)
	}

	skip := int64((page - 1) * pageSize)
	opts := options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(pageSize))

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("transaction findByUserID: %w", err)
	}
	defer cursor.Close(ctx)

	var txs []*models.Transaction
	if err := cursor.All(ctx, &txs); err != nil {
		return nil, 0, fmt.Errorf("transaction decode list: %w", err)
	}
	return txs, total, nil
}

func (r *mongoTransactionRepo) Update(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	tx.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"category_id": tx.CategoryID,
			"type":        tx.Type,
			"amount":      tx.Amount,
			"description": tx.Description,
			"date":        tx.Date,
			"updated_at":  tx.UpdatedAt,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.M{"_id": tx.ID, "user_id": tx.UserID}

	var result models.Transaction
	err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("transaction update: %w", err)
	}
	return &result, nil
}

// Delete removes a transaction only if it belongs to the given user.
func (r *mongoTransactionRepo) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	if err != nil {
		return fmt.Errorf("transaction delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// ExistsByCategoryID reports whether the user has any transactions referencing categoryID.
func (r *mongoTransactionRepo) ExistsByCategoryID(ctx context.Context, userID, categoryID primitive.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"user_id": userID, "category_id": categoryID})
	if err != nil {
		return false, fmt.Errorf("transaction existsByCategoryID: %w", err)
	}
	return count > 0, nil
}

// GetMonthlySummary aggregates inflow and outflow totals by calendar month in [since, until).
// A zero until means no upper bound.
func (r *mongoTransactionRepo) GetMonthlySummary(ctx context.Context, userID primitive.ObjectID, since, until time.Time) ([]*MonthlyAgg, error) {
	dateFilter := bson.M{"$gte": since}
	if !until.IsZero() {
		dateFilter["$lt"] = until
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"date":    dateFilter,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":  bson.M{"$year": "$date"},
				"month": bson.M{"$month": "$date"},
				"type":  "$type",
			},
			"total": bson.M{"$sum": "$amount"},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("GetMonthlySummary aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	type aggResult struct {
		ID struct {
			Year  int    `bson:"year"`
			Month int    `bson:"month"`
			Type  string `bson:"type"`
		} `bson:"_id"`
		Total float64 `bson:"total"`
	}

	// Merge results into MonthlyAgg map keyed by year*100+month.
	monthMap := make(map[int]*MonthlyAgg)
	for cursor.Next(ctx) {
		var doc aggResult
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("GetMonthlySummary decode: %w", err)
		}
		key := doc.ID.Year*100 + doc.ID.Month
		if _, ok := monthMap[key]; !ok {
			monthMap[key] = &MonthlyAgg{Year: doc.ID.Year, Month: doc.ID.Month}
		}
		switch doc.ID.Type {
		case "inflow":
			monthMap[key].Inflow += doc.Total
		case "outflow":
			monthMap[key].Outflow += doc.Total
		}
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetMonthlySummary cursor: %w", err)
	}

	result := make([]*MonthlyAgg, 0, len(monthMap))
	for _, agg := range monthMap {
		result = append(result, agg)
	}
	sort.Slice(result, func(i, j int) bool {
		ki := result[i].Year*100 + result[i].Month
		kj := result[j].Year*100 + result[j].Month
		return ki < kj
	})
	return result, nil
}

// GetCategoryTotals aggregates spending totals by category for the given type in [since, until),
// sorted descending by total. A zero until means no upper bound.
func (r *mongoTransactionRepo) GetCategoryTotals(ctx context.Context, userID primitive.ObjectID, txType string, since, until time.Time) ([]*CategoryAgg, error) {
	dateFilter := bson.M{"$gte": since}
	if !until.IsZero() {
		dateFilter["$lt"] = until
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"type":    txType,
			"date":    dateFilter,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$category_id",
			"total": bson.M{"$sum": "$amount"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("GetCategoryTotals aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	type aggResult struct {
		ID    primitive.ObjectID `bson:"_id"`
		Total float64            `bson:"total"`
	}

	var result []*CategoryAgg
	for cursor.Next(ctx) {
		var doc aggResult
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("GetCategoryTotals decode: %w", err)
		}
		result = append(result, &CategoryAgg{CategoryID: doc.ID, Total: doc.Total})
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetCategoryTotals cursor: %w", err)
	}
	return result, nil
}

// EnsureTransactionIndexes creates indexes for efficient query patterns.
func EnsureTransactionIndexes(ctx context.Context, db *mongo.Database) error {
	col := db.Collection(transactionsCollection)
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "date", Value: -1}}},
	})
	return err
}
