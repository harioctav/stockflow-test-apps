package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hari/stockflow/internal/model"
	"github.com/hari/stockflow/internal/store"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := strings.TrimSpace(q.Get("search"))
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	data, total, err := h.store.ListProducts(r.Context(), search, page, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	writeJSON(w, http.StatusOK, model.ProductListResp{
		Data:  data,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	req.SKU = strings.TrimSpace(req.SKU)
	req.Name = strings.TrimSpace(req.Name)
	if req.SKU == "" {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "sku is required")
		return
	}
	if req.Name == "" {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "name is required")
		return
	}
	if req.Price < 0 {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "price must be >= 0")
		return
	}
	if req.Stock != nil && *req.Stock < 0 {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "stock must be >= 0")
		return
	}

	p, err := h.store.CreateProduct(r.Context(), req)
	if errors.Is(err, store.ErrDuplicateSKU) {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "sku already exists")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	var req model.UpdateProductReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	req.SKU = strings.TrimSpace(req.SKU)
	req.Name = strings.TrimSpace(req.Name)
	if req.SKU == "" || req.Name == "" {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "sku and name are required")
		return
	}
	if req.Price < 0 {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "price must be >= 0")
		return
	}

	p, err := h.store.UpdateProduct(r.Context(), id, req)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "product not found")
		return
	}
	if errors.Is(err, store.ErrDuplicateSKU) {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "sku already exists")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	err = h.store.DeleteProduct(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "product not found")
		return
	}
	if errors.Is(err, store.ErrProductHasHistory) {
		writeJSON(w, http.StatusConflict, model.ErrorResp{
			Error: model.APIError{Code: "PRODUCT_HAS_HISTORY", Message: "product has movement history"},
		})
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	var req model.AdjustStockReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}
	if req.Type != "in" && req.Type != "out" {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "type must be 'in' or 'out'")
		return
	}
	if req.Qty <= 0 {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "qty must be > 0")
		return
	}

	p, err := h.store.AdjustStock(r.Context(), id, req.Type, req.Qty, req.Note)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "product not found")
		return
	}
	var stockErr *store.InsufficientStockError
	if errors.As(err, &stockErr) {
		writeStockErr(w, stockErr.SKU, stockErr.Available)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ListMovements(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	data, err := h.store.ListMovements(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "product not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}
