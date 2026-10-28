package model

import "time"

type Product struct {
	ID              int        `json:"id"`
	SKU             string     `json:"sku"`
	Name            string     `json:"name"`
	Price           int        `json:"price"`
	Stock           int        `json:"stock"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	LastMovementAt  *time.Time `json:"last_movement_at"`
}

type StockMovement struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Type      string    `json:"type"`
	Qty       int       `json:"qty"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductReq struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock *int   `json:"stock"`
}

type UpdateProductReq struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	IsActive *bool  `json:"is_active"`
}

type AdjustStockReq struct {
	Type string `json:"type"`
	Qty  int    `json:"qty"`
	Note string `json:"note"`
}

type OrderItemReq struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type CreateOrderReq struct {
	Customer string         `json:"customer"`
	Items    []OrderItemReq `json:"items"`
}

type OrderItem struct {
	ID        int    `json:"id"`
	OrderID   int    `json:"order_id"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}

type SalesOrder struct {
	ID          int         `json:"id"`
	Customer    string      `json:"customer"`
	Status      string      `json:"status"`
	Total       int         `json:"total"`
	CreatedAt   time.Time   `json:"created_at"`
	CancelledAt *time.Time  `json:"cancelled_at"`
	Items       []OrderItem `json:"items,omitempty"`
}

type ProductListResp struct {
	Data  []Product `json:"data"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Total int       `json:"total"`
}

type APIError struct {
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Available *int   `json:"available,omitempty"`
	SKU       string `json:"sku,omitempty"`
}

type ErrorResp struct {
	Error APIError `json:"error"`
}
