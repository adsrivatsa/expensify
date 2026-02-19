package services_test

import (
	"context"
	"testing"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"
	"expensify/internal/services"
	"expensify/internal/testutil"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newTxSvc(txRepo *testutil.MockTransactionRepo, catRepo *testutil.MockCategoryRepo) services.TransactionService {
	return services.NewTransactionService(txRepo, catRepo)
}

func TestTransactionService_Create(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	cat := &models.Category{ID: catID, Name: "Food", Icon: "🍕", Color: "#ff0000"}

	txRepo := &testutil.MockTransactionRepo{
		CreateFn: func(_ context.Context, tx *models.Transaction) (*models.Transaction, error) {
			tx.ID = primitive.NewObjectID()
			tx.CreatedAt = time.Now()
			tx.UpdatedAt = time.Now()
			return tx, nil
		},
	}
	catRepo := &testutil.MockCategoryRepo{
		FindByIDFn: func(_ context.Context, id primitive.ObjectID) (*models.Category, error) {
			if id == catID {
				return cat, nil
			}
			return nil, nil
		},
	}

	svc := newTxSvc(txRepo, catRepo)
	req := services.CreateTransactionRequest{
		CategoryID:  catID.Hex(),
		Type:        "outflow",
		Amount:      49.99,
		Description: "Dinner",
		Date:        time.Now(),
	}

	resp, err := svc.Create(context.Background(), userID.Hex(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.Amount != 49.99 {
		t.Errorf("amount: got %v, want 49.99", resp.Amount)
	}
	if resp.CategoryName != "Food" {
		t.Errorf("category name: got %q, want Food", resp.CategoryName)
	}
	if resp.Type != "outflow" {
		t.Errorf("type: got %q, want outflow", resp.Type)
	}
}

func TestTransactionService_Create_InvalidIDs(t *testing.T) {
	svc := newTxSvc(&testutil.MockTransactionRepo{}, &testutil.MockCategoryRepo{})

	_, err := svc.Create(context.Background(), "bad-uid", services.CreateTransactionRequest{CategoryID: primitive.NewObjectID().Hex(), Amount: 10})
	if err != services.ErrInvalidID {
		t.Errorf("expected ErrInvalidID for bad user ID, got %v", err)
	}

	uid := primitive.NewObjectID()
	_, err = svc.Create(context.Background(), uid.Hex(), services.CreateTransactionRequest{CategoryID: "bad-cat", Amount: 10})
	if err != services.ErrInvalidID {
		t.Errorf("expected ErrInvalidID for bad cat ID, got %v", err)
	}
}

func TestTransactionService_List_EnrichesWithCategory(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	txs := []*models.Transaction{
		{ID: primitive.NewObjectID(), UserID: userID, CategoryID: catID, Amount: 20, Date: time.Now()},
	}
	cats := []*models.Category{
		{ID: catID, Name: "Travel", Icon: "✈️", Color: "#4ecdc4"},
	}

	txRepo := &testutil.MockTransactionRepo{
		FindByUserIDFn: func(_ context.Context, _ primitive.ObjectID, _, _ int) ([]*models.Transaction, int64, error) {
			return txs, 1, nil
		},
	}
	catRepo := &testutil.MockCategoryRepo{
		FindByIDsFn: func(_ context.Context, _ []primitive.ObjectID) ([]*models.Category, error) {
			return cats, nil
		},
	}

	svc := newTxSvc(txRepo, catRepo)
	result, err := svc.List(context.Background(), userID.Hex(), 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	item := result.Items[0]
	if item.CategoryName != "Travel" {
		t.Errorf("category name: got %q, want Travel", item.CategoryName)
	}
	if item.CategoryIcon != "✈️" {
		t.Errorf("category icon: got %q, want ✈️", item.CategoryIcon)
	}
}

func TestTransactionService_List_Pagination(t *testing.T) {
	userID := primitive.NewObjectID()

	txRepo := &testutil.MockTransactionRepo{
		FindByUserIDFn: func(_ context.Context, _ primitive.ObjectID, page, pageSize int) ([]*models.Transaction, int64, error) {
			return []*models.Transaction{}, 47, nil
		},
	}
	catRepo := &testutil.MockCategoryRepo{}

	svc := newTxSvc(txRepo, catRepo)
	result, err := svc.List(context.Background(), userID.Hex(), 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// ceil(47/20) = 3
	if result.TotalPages != 3 {
		t.Errorf("total_pages: got %d, want 3", result.TotalPages)
	}
}

func TestTransactionService_Update_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	txID := primitive.NewObjectID()

	updatedTx := &models.Transaction{
		ID: txID, UserID: userID, CategoryID: catID, Amount: 75, Description: "updated", Date: time.Now(),
	}
	cat := &models.Category{ID: catID, Name: "Shopping", Icon: "🛍️", Color: "#45b7d1"}

	txRepo := &testutil.MockTransactionRepo{
		UpdateFn: func(_ context.Context, tx *models.Transaction) (*models.Transaction, error) {
			return updatedTx, nil
		},
	}
	catRepo := &testutil.MockCategoryRepo{
		FindByIDFn: func(_ context.Context, _ primitive.ObjectID) (*models.Category, error) { return cat, nil },
	}

	svc := newTxSvc(txRepo, catRepo)
	req := services.UpdateTransactionRequest{
		CategoryID: catID.Hex(), Amount: 75, Description: "updated", Date: time.Now(),
	}
	resp, err := svc.Update(context.Background(), userID.Hex(), txID.Hex(), req)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if resp.Amount != 75 {
		t.Errorf("amount: got %v, want 75", resp.Amount)
	}
}

func TestTransactionService_Update_NotOwned(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	txID := primitive.NewObjectID()

	txRepo := &testutil.MockTransactionRepo{
		UpdateFn: func(_ context.Context, _ *models.Transaction) (*models.Transaction, error) { return nil, db.ErrNotFound },
	}

	svc := newTxSvc(txRepo, &testutil.MockCategoryRepo{})
	_, err := svc.Update(context.Background(), userID.Hex(), txID.Hex(), services.UpdateTransactionRequest{
		CategoryID: catID.Hex(), Amount: 10,
	})
	if err != services.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Delete_Success(t *testing.T) {
	userID := primitive.NewObjectID()
	txID := primitive.NewObjectID()
	deleted := false

	txRepo := &testutil.MockTransactionRepo{
		DeleteFn: func(_ context.Context, id, uid primitive.ObjectID) error {
			if id == txID && uid == userID {
				deleted = true
			}
			return nil
		},
	}

	svc := newTxSvc(txRepo, &testutil.MockCategoryRepo{})
	if err := svc.Delete(context.Background(), userID.Hex(), txID.Hex()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !deleted {
		t.Error("expected repo.Delete to be called")
	}
}

func TestTransactionService_Delete_NotOwned(t *testing.T) {
	userID := primitive.NewObjectID()
	txID := primitive.NewObjectID()

	txRepo := &testutil.MockTransactionRepo{
		DeleteFn: func(_ context.Context, _, _ primitive.ObjectID) error { return db.ErrNotFound },
	}

	svc := newTxSvc(txRepo, &testutil.MockCategoryRepo{})
	err := svc.Delete(context.Background(), userID.Hex(), txID.Hex())
	if err != services.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionService_Summary(t *testing.T) {
	userID := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	txRepo := &testutil.MockTransactionRepo{
		GetMonthlySummaryFn: func(_ context.Context, _ primitive.ObjectID, _, _ time.Time) ([]*db.MonthlyAgg, error) {
			return []*db.MonthlyAgg{
				{Year: 2024, Month: 1, Inflow: 1000, Outflow: 500},
				{Year: 2024, Month: 2, Inflow: 0, Outflow: 300},
			}, nil
		},
		GetCategoryTotalsFn: func(_ context.Context, _ primitive.ObjectID, _ string, _, _ time.Time) ([]*db.CategoryAgg, error) {
			return []*db.CategoryAgg{
				{CategoryID: catID, Total: 500},
			}, nil
		},
	}
	catRepo := &testutil.MockCategoryRepo{
		FindByIDsFn: func(_ context.Context, _ []primitive.ObjectID) ([]*models.Category, error) {
			return []*models.Category{
				{ID: catID, Name: "Food", Icon: "🍕", Color: "#ff0000"},
			}, nil
		},
	}

	svc := newTxSvc(txRepo, catRepo)
	since := time.Now().AddDate(0, -6, 0)
	summary, err := svc.Summary(context.Background(), userID.Hex(), since, time.Time{})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if len(summary.Monthly) != 2 {
		t.Errorf("monthly: got %d, want 2", len(summary.Monthly))
	}
	if summary.Monthly[0].Inflow != 1000 {
		t.Errorf("monthly[0].Inflow: got %v, want 1000", summary.Monthly[0].Inflow)
	}
	if len(summary.ByCategory) != 1 {
		t.Fatalf("by_category: got %d, want 1", len(summary.ByCategory))
	}
	if summary.ByCategory[0].CategoryName != "Food" {
		t.Errorf("category name: got %q, want Food", summary.ByCategory[0].CategoryName)
	}
	if summary.ByCategory[0].Total != 500 {
		t.Errorf("category total: got %v, want 500", summary.ByCategory[0].Total)
	}
}
