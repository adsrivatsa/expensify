package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"expensify/internal/db"
	"expensify/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateTransactionRequest holds the fields for a new transaction.
type CreateTransactionRequest struct {
	CategoryID  string    `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// UpdateTransactionRequest holds updatable transaction fields.
type UpdateTransactionRequest struct {
	CategoryID  string    `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// TransactionResponse is the enriched view of a transaction returned to clients.
type TransactionResponse struct {
	ID            string    `json:"id"`
	CategoryID    string    `json:"category_id"`
	CategoryName  string    `json:"category_name"`
	CategoryColor string    `json:"category_color"`
	CategoryIcon  string    `json:"category_icon"`
	Type          string    `json:"type"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	Date          time.Time `json:"date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MonthlyPoint holds aggregated cashflow totals for a single month.
type MonthlyPoint struct {
	Year    int     `json:"year"`
	Month   int     `json:"month"`
	Inflow  float64 `json:"inflow"`
	Outflow float64 `json:"outflow"`
}

// CategoryPoint holds outflow totals for a category, enriched with category metadata.
type CategoryPoint struct {
	CategoryID    string  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	CategoryIcon  string  `json:"category_icon"`
	Total         float64 `json:"total"`
}

// CashflowSummary is the response for the summary endpoint.
type CashflowSummary struct {
	Monthly    []*MonthlyPoint  `json:"monthly"`
	ByCategory []*CategoryPoint `json:"by_category"`
}

// PaginatedTransactions wraps a page of transaction responses with metadata.
type PaginatedTransactions struct {
	Items      []*TransactionResponse `json:"items"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// TransactionService manages spending transactions.
type TransactionService interface {
	Create(ctx context.Context, userID string, req CreateTransactionRequest) (*TransactionResponse, error)
	List(ctx context.Context, userID string, page, pageSize int) (*PaginatedTransactions, error)
	Update(ctx context.Context, userID string, txID string, req UpdateTransactionRequest) (*TransactionResponse, error)
	Delete(ctx context.Context, userID string, txID string) error
	Summary(ctx context.Context, userID string, since, until time.Time) (*CashflowSummary, error)
}

type transactionService struct {
	txRepo  db.TransactionRepository
	catRepo db.CategoryRepository
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(txRepo db.TransactionRepository, catRepo db.CategoryRepository) TransactionService {
	return &transactionService{txRepo: txRepo, catRepo: catRepo}
}

func (s *transactionService) Create(ctx context.Context, userID string, req CreateTransactionRequest) (*TransactionResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}
	catID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, ErrInvalidID
	}

	tx := &models.Transaction{
		UserID:      uid,
		CategoryID:  catID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
	}
	created, err := s.txRepo.Create(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("creating transaction: %w", err)
	}

	cat, _ := s.catRepo.FindByID(ctx, catID)
	return toResponse(created, cat), nil
}

func (s *transactionService) List(ctx context.Context, userID string, page, pageSize int) (*PaginatedTransactions, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}

	txs, total, err := s.txRepo.FindByUserID(ctx, uid, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("fetching transactions: %w", err)
	}

	// Collect unique category IDs for a single batch fetch.
	seen := make(map[primitive.ObjectID]struct{})
	for _, tx := range txs {
		seen[tx.CategoryID] = struct{}{}
	}
	ids := make([]primitive.ObjectID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	cats := make(map[primitive.ObjectID]*models.Category)
	if len(ids) > 0 {
		fetched, err := s.catRepo.FindByIDs(ctx, ids)
		if err == nil {
			for _, c := range fetched {
				cats[c.ID] = c
			}
		}
	}

	responses := make([]*TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = toResponse(tx, cats[tx.CategoryID])
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &PaginatedTransactions{
		Items:      responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *transactionService) Update(ctx context.Context, userID string, txID string, req UpdateTransactionRequest) (*TransactionResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}
	tid, err := primitive.ObjectIDFromHex(txID)
	if err != nil {
		return nil, ErrInvalidID
	}
	catID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, ErrInvalidID
	}

	tx := &models.Transaction{
		ID:          tid,
		UserID:      uid,
		CategoryID:  catID,
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
	}
	updated, err := s.txRepo.Update(ctx, tx)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("updating transaction: %w", err)
	}

	cat, _ := s.catRepo.FindByID(ctx, catID)
	return toResponse(updated, cat), nil
}

func (s *transactionService) Delete(ctx context.Context, userID string, txID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrInvalidID
	}
	tid, err := primitive.ObjectIDFromHex(txID)
	if err != nil {
		return ErrInvalidID
	}

	if err := s.txRepo.Delete(ctx, tid, uid); err != nil {
		if err == db.ErrNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("deleting transaction: %w", err)
	}
	return nil
}

func toResponse(tx *models.Transaction, cat *models.Category) *TransactionResponse {
	resp := &TransactionResponse{
		ID:          tx.ID.Hex(),
		CategoryID:  tx.CategoryID.Hex(),
		Type:        tx.Type,
		Amount:      tx.Amount,
		Description: tx.Description,
		Date:        tx.Date,
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
	if cat != nil {
		resp.CategoryName = cat.Name
		resp.CategoryColor = cat.Color
		resp.CategoryIcon = cat.Icon
	}
	return resp
}

func (s *transactionService) Summary(ctx context.Context, userID string, since, until time.Time) (*CashflowSummary, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrInvalidID
	}

	monthlyAggs, err := s.txRepo.GetMonthlySummary(ctx, uid, since, until)
	if err != nil {
		return nil, fmt.Errorf("monthly summary: %w", err)
	}

	catAggs, err := s.txRepo.GetCategoryTotals(ctx, uid, "outflow", since, until)
	if err != nil {
		return nil, fmt.Errorf("category totals: %w", err)
	}

	// Batch-fetch categories for enrichment.
	catIDs := make([]primitive.ObjectID, 0, len(catAggs))
	for _, ca := range catAggs {
		catIDs = append(catIDs, ca.CategoryID)
	}
	catMap := make(map[primitive.ObjectID]*models.Category)
	if len(catIDs) > 0 {
		fetched, err := s.catRepo.FindByIDs(ctx, catIDs)
		if err == nil {
			for _, c := range fetched {
				catMap[c.ID] = c
			}
		}
	}

	monthly := make([]*MonthlyPoint, len(monthlyAggs))
	for i, a := range monthlyAggs {
		monthly[i] = &MonthlyPoint{Year: a.Year, Month: a.Month, Inflow: a.Inflow, Outflow: a.Outflow}
	}

	byCategory := make([]*CategoryPoint, 0, len(catAggs))
	for _, ca := range catAggs {
		cp := &CategoryPoint{
			CategoryID: ca.CategoryID.Hex(),
			Total:      ca.Total,
		}
		if cat, ok := catMap[ca.CategoryID]; ok {
			cp.CategoryName = cat.Name
			cp.CategoryColor = cat.Color
			cp.CategoryIcon = cat.Icon
		}
		byCategory = append(byCategory, cp)
	}

	return &CashflowSummary{Monthly: monthly, ByCategory: byCategory}, nil
}
