//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func makeTransaction(userID, catID primitive.ObjectID, amount float64, date time.Time) *models.Transaction {
	return &models.Transaction{
		UserID:      userID,
		CategoryID:  catID,
		Type:        "outflow",
		Amount:      amount,
		Description: "test tx",
		Date:        date,
	}
}

func TestTransactionRepo_Create(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	tx := makeTransaction(uid, catID, 42.50, time.Now())

	created, err := repo.Create(ctx, tx)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID.IsZero() {
		t.Error("expected non-zero transaction ID")
	}
	if created.Amount != 42.50 {
		t.Errorf("amount: got %v, want 42.50", created.Amount)
	}
}

func TestTransactionRepo_FindByID(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	created, _ := repo.Create(ctx, makeTransaction(uid, catID, 10, time.Now()))

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Error("FindByID returned wrong or nil result")
	}

	missing, err := repo.FindByID(ctx, primitive.NewObjectID())
	if err != nil {
		t.Fatalf("FindByID missing: %v", err)
	}
	if missing != nil {
		t.Error("expected nil for missing transaction")
	}
}

func TestTransactionRepo_FindByUserID_Pagination(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	// Insert 5 transactions on different days.
	for i := 0; i < 5; i++ {
		repo.Create(ctx, makeTransaction(uid, catID, float64(i+1)*10, time.Now().Add(time.Duration(-i)*24*time.Hour)))
	}

	// Page 1, size 3 → 3 items, total 5.
	page1, total, err := repo.FindByUserID(ctx, uid, 1, 3)
	if err != nil {
		t.Fatalf("FindByUserID page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("total: got %d, want 5", total)
	}
	if len(page1) != 3 {
		t.Errorf("page1 items: got %d, want 3", len(page1))
	}

	// Page 2, size 3 → 2 items.
	page2, _, err := repo.FindByUserID(ctx, uid, 2, 3)
	if err != nil {
		t.Fatalf("FindByUserID page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("page2 items: got %d, want 2", len(page2))
	}

	// Results should be newest-first.
	if !page1[0].Date.After(page1[1].Date) {
		t.Error("expected results sorted newest-first")
	}
}

func TestTransactionRepo_FindByUserID_OtherUserIsolation(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid1 := primitive.NewObjectID()
	uid2 := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	repo.Create(ctx, makeTransaction(uid1, catID, 100, time.Now()))
	repo.Create(ctx, makeTransaction(uid2, catID, 200, time.Now()))

	txs, total, _ := repo.FindByUserID(ctx, uid1, 1, 20)
	if total != 1 || len(txs) != 1 {
		t.Errorf("user isolation failed: got %d transactions for uid1", len(txs))
	}
}

func TestTransactionRepo_Update(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	created, _ := repo.Create(ctx, makeTransaction(uid, catID, 50, time.Now()))

	created.Amount = 99.99
	created.Description = "updated"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Amount != 99.99 {
		t.Errorf("amount: got %v, want 99.99", updated.Amount)
	}
	if updated.Description != "updated" {
		t.Errorf("description: got %q, want updated", updated.Description)
	}
}

func TestTransactionRepo_Update_WrongUser(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	created, _ := repo.Create(ctx, makeTransaction(ownerID, catID, 50, time.Now()))
	created.UserID = otherID // Impersonate another user.

	_, err := repo.Update(ctx, created)
	if err != db.ErrNotFound {
		t.Errorf("expected ErrNotFound when updating another user's transaction, got %v", err)
	}
}

func TestTransactionRepo_Delete(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	created, _ := repo.Create(ctx, makeTransaction(uid, catID, 30, time.Now()))

	if err := repo.Delete(ctx, created.ID, uid); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	found, _ := repo.FindByID(ctx, created.ID)
	if found != nil {
		t.Error("expected transaction to be deleted")
	}
}

func TestTransactionRepo_Delete_WrongUser(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	ownerID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	catID := primitive.NewObjectID()
	created, _ := repo.Create(ctx, makeTransaction(ownerID, catID, 30, time.Now()))

	err := repo.Delete(ctx, created.ID, otherID)
	if err != db.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTransactionRepo_GetMonthlySummary(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catID := primitive.NewObjectID()

	// Insert inflow and outflow across two months.
	jan := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)

	inflow := makeTransaction(uid, catID, 500, jan)
	inflow.Type = "inflow"
	repo.Create(ctx, inflow)

	outflow := makeTransaction(uid, catID, 200, jan)
	outflow.Type = "outflow"
	repo.Create(ctx, outflow)

	outflow2 := makeTransaction(uid, catID, 300, feb)
	outflow2.Type = "outflow"
	repo.Create(ctx, outflow2)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	aggs, err := repo.GetMonthlySummary(ctx, uid, since, time.Time{})
	if err != nil {
		t.Fatalf("GetMonthlySummary: %v", err)
	}
	if len(aggs) != 2 {
		t.Fatalf("expected 2 months, got %d", len(aggs))
	}
	// Results sorted by year/month ascending.
	janAgg := aggs[0]
	if janAgg.Month != 1 || janAgg.Inflow != 500 || janAgg.Outflow != 200 {
		t.Errorf("jan agg mismatch: %+v", janAgg)
	}
	febAgg := aggs[1]
	if febAgg.Month != 2 || febAgg.Outflow != 300 {
		t.Errorf("feb agg mismatch: %+v", febAgg)
	}
}

func TestTransactionRepo_GetCategoryTotals(t *testing.T) {
	repo := db.NewTransactionRepository(testDB(t))
	ctx := context.Background()

	uid := primitive.NewObjectID()
	catA := primitive.NewObjectID()
	catB := primitive.NewObjectID()

	now := time.Now()
	// catA gets 150, catB gets 300.
	tx1 := makeTransaction(uid, catA, 100, now)
	tx1.Type = "outflow"
	repo.Create(ctx, tx1)

	tx2 := makeTransaction(uid, catA, 50, now)
	tx2.Type = "outflow"
	repo.Create(ctx, tx2)

	tx3 := makeTransaction(uid, catB, 300, now)
	tx3.Type = "outflow"
	repo.Create(ctx, tx3)

	// Inflow should be excluded.
	tx4 := makeTransaction(uid, catA, 999, now)
	tx4.Type = "inflow"
	repo.Create(ctx, tx4)

	since := now.AddDate(0, -1, 0)
	aggs, err := repo.GetCategoryTotals(ctx, uid, "outflow", since, time.Time{})
	if err != nil {
		t.Fatalf("GetCategoryTotals: %v", err)
	}
	if len(aggs) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(aggs))
	}
	// Sorted descending by total: catB (300) first.
	if aggs[0].CategoryID != catB || aggs[0].Total != 300 {
		t.Errorf("first agg mismatch: %+v", aggs[0])
	}
	if aggs[1].CategoryID != catA || aggs[1].Total != 150 {
		t.Errorf("second agg mismatch: %+v", aggs[1])
	}
}
