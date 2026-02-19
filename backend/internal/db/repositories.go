package db

import (
	"context"
	"time"

	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MonthlyAgg holds aggregated inflow/outflow totals for a single calendar month.
type MonthlyAgg struct {
	Year    int
	Month   int
	Inflow  float64
	Outflow float64
}

// CategoryAgg holds the total outflow for a single category.
type CategoryAgg struct {
	CategoryID primitive.ObjectID
	Total      float64
}

// UserRepository defines persistence operations for users.
type UserRepository interface {
	FindByGoogleID(ctx context.Context, googleID string) (*models.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	Upsert(ctx context.Context, user *models.User) (*models.User, error)
}

// SessionRepository defines persistence operations for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	FindByToken(ctx context.Context, token string) (*models.Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
	FindDefaultCategories(ctx context.Context) ([]*models.Category, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Category, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Category, error)
	Create(ctx context.Context, category *models.Category) (*models.Category, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) (*models.Transaction, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*models.Transaction, int64, error)
	Update(ctx context.Context, tx *models.Transaction) (*models.Transaction, error)
	Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	ExistsByCategoryID(ctx context.Context, userID primitive.ObjectID, categoryID primitive.ObjectID) (bool, error)
	GetMonthlySummary(ctx context.Context, userID primitive.ObjectID, since, until time.Time) ([]*MonthlyAgg, error)
	GetCategoryTotals(ctx context.Context, userID primitive.ObjectID, txType string, since, until time.Time) ([]*CategoryAgg, error)
}
