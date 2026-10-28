package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hari/stockflow/internal/model"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, model.ErrorResp{
		Error: model.APIError{Code: code, Message: msg},
	})
}

func writeStockErr(w http.ResponseWriter, sku string, available int) {
	avail := available
	writeJSON(w, http.StatusConflict, model.ErrorResp{
		Error: model.APIError{
			Code:      "INSUFFICIENT_STOCK",
			Message:   "insufficient stock",
			Available: &avail,
			SKU:       sku,
		},
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
