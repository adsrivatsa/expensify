package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"expensify/internal/middleware"
	"expensify/internal/services"

	"github.com/go-chi/chi/v5"
)

// CategoryHandler handles CRUD for spending categories.
type CategoryHandler struct {
	svc services.CategoryService
}

// NewCategoryHandler constructs a CategoryHandler.
func NewCategoryHandler(svc services.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List returns all categories available to the user (defaults + custom).
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	cats, err := h.svc.GetCategories(r.Context(), user.ID.Hex())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

// Create adds a new custom category for the authenticated user.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var req services.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), user.ID.Hex(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create category")
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

// Delete removes a custom category owned by the authenticated user.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	catID := chi.URLParam(r, "id")

	if err := h.svc.DeleteCategory(r.Context(), user.ID.Hex(), catID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found or not owned by you")
			return
		}
		if errors.Is(err, services.ErrInvalidID) {
			writeError(w, http.StatusBadRequest, "invalid category id")
			return
		}
		if errors.Is(err, services.ErrCategoryInUse) {
			writeError(w, http.StatusConflict, "category has existing transactions and cannot be deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete category")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
