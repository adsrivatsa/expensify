package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"expensify/internal/middleware"
	"expensify/internal/services"

	"github.com/go-chi/chi/v5"
)

// TransactionHandler handles CRUD for spending transactions.
type TransactionHandler struct {
	svc services.TransactionService
}

// NewTransactionHandler constructs a TransactionHandler.
func NewTransactionHandler(svc services.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// List returns a paginated list of transactions for the authenticated user.
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.svc.List(r.Context(), user.ID.Hex(), page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch transactions")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// Create adds a new transaction for the authenticated user.
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var req services.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Amount == 0 || req.CategoryID == "" {
		writeError(w, http.StatusBadRequest, "amount and category_id are required")
		return
	}

	tx, err := h.svc.Create(r.Context(), user.ID.Hex(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidID) {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create transaction")
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

// Update modifies an existing transaction owned by the authenticated user.
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	txID := chi.URLParam(r, "id")

	var req services.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Update(r.Context(), user.ID.Hex(), txID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			writeError(w, http.StatusNotFound, "transaction not found")
		case errors.Is(err, services.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid id")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update transaction")
		}
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// Delete removes a transaction owned by the authenticated user.
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	txID := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), user.ID.Hex(), txID); err != nil {
		switch {
		case errors.Is(err, services.ErrNotFound):
			writeError(w, http.StatusNotFound, "transaction not found")
		case errors.Is(err, services.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid id")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete transaction")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Summary returns aggregated cashflow data for the authenticated user.
// Accepts ?year=YYYY for a calendar year view, or ?months=N for a trailing window (default 12).
func (h *TransactionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var since, until time.Time

	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 2000 || year > 2100 {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}
		since = time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		until = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	} else {
		months := queryInt(r, "months", 12)
		if months < 1 {
			months = 1
		}
		if months > 24 {
			months = 24
		}
		since = time.Now().AddDate(0, -months, 0)
		// until is zero — no upper bound, shows up to now
	}

	summary, err := h.svc.Summary(r.Context(), user.ID.Hex(), since, until)
	if err != nil {
		if errors.Is(err, services.ErrInvalidID) {
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch summary")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
