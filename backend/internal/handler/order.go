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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderReq
	if err := decodeJSON(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json body")
		return
	}

	req.Customer = strings.TrimSpace(req.Customer)
	if req.Customer == "" {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "customer is required")
		return
	}
	if len(req.Items) == 0 {
		writeErr(w, http.StatusBadRequest, "VALIDATION", "items cannot be empty")
		return
	}

	seen := map[string]bool{}
	for i, it := range req.Items {
		sku := strings.TrimSpace(it.SKU)
		if sku == "" {
			writeErr(w, http.StatusBadRequest, "VALIDATION", "item sku is required")
			return
		}
		if it.Qty <= 0 {
			writeErr(w, http.StatusBadRequest, "VALIDATION", "item qty must be > 0")
			return
		}
		if seen[sku] {
			writeErr(w, http.StatusBadRequest, "VALIDATION", "duplicate sku in order: "+sku)
			return
		}
		seen[sku] = true
		req.Items[i].SKU = sku
	}

	order, err := h.store.CreateOrder(r.Context(), req.Customer, req.Items)
	var stockErr *store.InsufficientStockError
	if errors.As(err, &stockErr) {
		writeStockErr(w, stockErr.SKU, stockErr.Available)
		return
	}
	var valErr *store.ValidationError
	if errors.As(err, &valErr) {
		writeErr(w, http.StatusBadRequest, "VALIDATION", valErr.Msg)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	order, err := h.store.GetOrder(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "order not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid id")
		return
	}
	order, err := h.store.CancelOrder(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "order not found")
		return
	}
	if errors.Is(err, store.ErrAlreadyCancelled) {
		writeErr(w, http.StatusConflict, "ALREADY_CANCELLED", "order already cancelled")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}
