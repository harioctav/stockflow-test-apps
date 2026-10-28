package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hari/stockflow/internal/model"
	"github.com/jackc/pgx/v5"
)

type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

func (s *Store) CreateOrder(ctx context.Context, customer string, items []model.OrderItemReq) (*model.SalesOrder, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	type line struct {
		sku       string
		qty       int
		unitPrice int
		productID int
	}

	type locked struct {
		id     int
		sku    string
		price  int
		stock  int
		active bool
		qty    int
	}

	lockedRows := make([]locked, 0, len(items))
	for _, it := range items {
		sku := strings.TrimSpace(it.SKU)
		var row locked
		err := tx.QueryRow(ctx, `
			SELECT id, sku, price, stock, is_active
			FROM products WHERE sku = $1
			FOR UPDATE
		`, sku).Scan(&row.id, &row.sku, &row.price, &row.stock, &row.active)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &ValidationError{Msg: fmt.Sprintf("unknown sku: %s", sku)}
		}
		if err != nil {
			return nil, err
		}
		if !row.active {
			return nil, &ValidationError{Msg: fmt.Sprintf("product %s is inactive", sku)}
		}
		row.qty = it.Qty
		lockedRows = append(lockedRows, row)
	}

	total := 0
	lines := make([]line, 0, len(lockedRows))
	for _, row := range lockedRows {
		ct, err := tx.Exec(ctx, `
			UPDATE products SET stock = stock - $2
			WHERE id = $1 AND stock >= $2
		`, row.id, row.qty)
		if err != nil {
			return nil, err
		}
		if ct.RowsAffected() == 0 {
			return nil, &InsufficientStockError{SKU: row.sku, Available: row.stock}
		}
		lines = append(lines, line{sku: row.sku, qty: row.qty, unitPrice: row.price, productID: row.id})
		total += row.price * row.qty
	}

	var order model.SalesOrder
	err = tx.QueryRow(ctx, `
		INSERT INTO sales_orders (customer, total)
		VALUES ($1, $2)
		RETURNING id, customer, status, total, created_at, cancelled_at
	`, strings.TrimSpace(customer), total).
		Scan(&order.ID, &order.Customer, &order.Status, &order.Total, &order.CreatedAt, &order.CancelledAt)
	if err != nil {
		return nil, err
	}

	order.Items = make([]model.OrderItem, 0, len(lines))
	for _, l := range lines {
		note := fmt.Sprintf("SO #%d", order.ID)
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_movements (product_id, type, qty, note)
			VALUES ($1, 'out', $2, $3)
		`, l.productID, l.qty, note)
		if err != nil {
			return nil, err
		}

		var oi model.OrderItem
		err = tx.QueryRow(ctx, `
			INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
			VALUES ($1, $2, $3, $4)
			RETURNING id, order_id, sku, qty, unit_price
		`, order.ID, l.sku, l.qty, l.unitPrice).
			Scan(&oi.ID, &oi.OrderID, &oi.SKU, &oi.Qty, &oi.UnitPrice)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *Store) GetOrder(ctx context.Context, id int) (*model.SalesOrder, error) {
	var o model.SalesOrder
	err := s.pool.QueryRow(ctx, `
		SELECT id, customer, status, total, created_at, cancelled_at
		FROM sales_orders WHERE id = $1
	`, id).Scan(&o.ID, &o.Customer, &o.Status, &o.Total, &o.CreatedAt, &o.CancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, order_id, sku, qty, unit_price
		FROM sales_order_items WHERE order_id = $1 ORDER BY id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	o.Items = []model.OrderItem{}
	for rows.Next() {
		var it model.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.SKU, &it.Qty, &it.UnitPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}

func (s *Store) CancelOrder(ctx context.Context, id int) (*model.SalesOrder, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
		SELECT status FROM sales_orders WHERE id = $1 FOR UPDATE
	`, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == "CANCELLED" {
		return nil, ErrAlreadyCancelled
	}

	rows, err := tx.Query(ctx, `
		SELECT sku, qty FROM sales_order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	type restore struct {
		sku string
		qty int
	}
	var items []restore
	for rows.Next() {
		var r restore
		if err := rows.Scan(&r.sku, &r.qty); err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	now := time.Now()
	for _, it := range items {
		var pid int
		err := tx.QueryRow(ctx, `SELECT id FROM products WHERE sku = $1 FOR UPDATE`, it.sku).Scan(&pid)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `UPDATE products SET stock = stock + $2 WHERE id = $1`, pid, it.qty)
		if err != nil {
			return nil, err
		}
		note := fmt.Sprintf("Cancel SO #%d", id)
		_, err = tx.Exec(ctx, `
			INSERT INTO stock_movements (product_id, type, qty, note)
			VALUES ($1, 'in', $2, $3)
		`, pid, it.qty, note)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE sales_orders SET status = 'CANCELLED', cancelled_at = $2 WHERE id = $1
	`, id, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetOrder(ctx, id)
}
