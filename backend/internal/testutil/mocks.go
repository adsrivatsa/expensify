// Package testutil provides hand-written mock implementations of all repository
// interfaces for use in service-layer unit tests.
package testutil

import (
	"context"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- UserRepository mock ----

type MockUserRepo struct {
	FindByGoogleIDFn func(ctx context.Context, googleID string) (*models.User, error)
	FindByIDFn       func(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	UpsertFn         func(ctx context.Context, user *models.User) (*models.User, error)
}

func (m *MockUserRepo) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	if m.FindByGoogleIDFn != nil {
		return m.FindByGoogleIDFn(ctx, googleID)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepo) Upsert(ctx context.Context, user *models.User) (*models.User, error) {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, user)
	}
	return nil, nil
}

// ---- SessionRepository mock ----

type MockSessionRepo struct {
	CreateFn        func(ctx context.Context, session *models.Session) (*models.Session, error)
	FindByTokenFn   func(ctx context.Context, token string) (*models.Session, error)
	DeleteFn        func(ctx context.Context, token string) error
	DeleteExpiredFn func(ctx context.Context) error
}

func (m *MockSessionRepo) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, session)
	}
	return nil, nil
}

func (m *MockSessionRepo) FindByToken(ctx context.Context, token string) (*models.Session, error) {
	if m.FindByTokenFn != nil {
		return m.FindByTokenFn(ctx, token)
	}
	return nil, nil
}

func (m *MockSessionRepo) Delete(ctx context.Context, token string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, token)
	}
	return nil
}

func (m *MockSessionRepo) DeleteExpired(ctx context.Context) error {
	if m.DeleteExpiredFn != nil {
		return m.DeleteExpiredFn(ctx)
	}
	return nil
}

// ---- CategoryRepository mock ----

type MockCategoryRepo struct {
	FindDefaultCategoriesFn func(ctx context.Context) ([]*models.Category, error)
	FindByUserIDFn          func(ctx context.Context, userID primitive.ObjectID) ([]*models.Category, error)
	FindByIDFn              func(ctx context.Context, id primitive.ObjectID) (*models.Category, error)
	FindByIDsFn             func(ctx context.Context, ids []primitive.ObjectID) ([]*models.Category, error)
	CreateFn                func(ctx context.Context, category *models.Category) (*models.Category, error)
	DeleteFn                func(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
}

func (m *MockCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*models.Category, error) {
	if m.FindDefaultCategoriesFn != nil {
		return m.FindDefaultCategoriesFn(ctx)
	}
	return nil, nil
}

func (m *MockCategoryRepo) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Category, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockCategoryRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Category, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockCategoryRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*models.Category, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *MockCategoryRepo) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, category)
	}
	return nil, nil
}

func (m *MockCategoryRepo) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, userID)
	}
	return nil
}

// ---- TransactionRepository mock ----

type MockTransactionRepo struct {
	CreateFn                func(ctx context.Context, tx *models.Transaction) (*models.Transaction, error)
	FindByIDFn              func(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error)
	FindByUserIDFn          func(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*models.Transaction, int64, error)
	UpdateFn                func(ctx context.Context, tx *models.Transaction) (*models.Transaction, error)
	DeleteFn                func(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
	ExistsByCategoryIDFn    func(ctx context.Context, userID primitive.ObjectID, categoryID primitive.ObjectID) (bool, error)
	GetMonthlySummaryFn     func(ctx context.Context, userID primitive.ObjectID, since, until time.Time) ([]*db.MonthlyAgg, error)
	GetCategoryTotalsFn     func(ctx context.Context, userID primitive.ObjectID, txType string, since, until time.Time) ([]*db.CategoryAgg, error)
}

func (m *MockTransactionRepo) Create(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tx)
	}
	return nil, nil
}

func (m *MockTransactionRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Transaction, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockTransactionRepo) FindByUserID(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*models.Transaction, int64, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockTransactionRepo) Update(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tx)
	}
	return nil, nil
}

func (m *MockTransactionRepo) Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, userID)
	}
	return nil
}

func (m *MockTransactionRepo) ExistsByCategoryID(ctx context.Context, userID primitive.ObjectID, categoryID primitive.ObjectID) (bool, error) {
	if m.ExistsByCategoryIDFn != nil {
		return m.ExistsByCategoryIDFn(ctx, userID, categoryID)
	}
	return false, nil
}

func (m *MockTransactionRepo) GetMonthlySummary(ctx context.Context, userID primitive.ObjectID, since, until time.Time) ([]*db.MonthlyAgg, error) {
	if m.GetMonthlySummaryFn != nil {
		return m.GetMonthlySummaryFn(ctx, userID, since, until)
	}
	return nil, nil
}

func (m *MockTransactionRepo) GetCategoryTotals(ctx context.Context, userID primitive.ObjectID, txType string, since, until time.Time) ([]*db.CategoryAgg, error) {
	if m.GetCategoryTotalsFn != nil {
		return m.GetCategoryTotalsFn(ctx, userID, txType, since, until)
	}
	return nil, nil
}
