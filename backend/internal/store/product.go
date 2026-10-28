package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hari/stockflow/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrProductHasHistory  = errors.New("product has history")
	ErrAlreadyCancelled   = errors.New("already cancelled")
	ErrDuplicateSKU       = errors.New("duplicate sku")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

type InsufficientStockError struct {
	SKU       string
	Available int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for %s (available %d)", e.SKU, e.Available)
}

func (e *InsufficientStockError) Unwrap() error {
	return ErrInsufficientStock
}

// last_movement_at comes from stock_movements (subquery), not a column on products
func (s *Store) ListProducts(ctx context.Context, search string, page, limit int) ([]model.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []any{}
	n := 1
	if search != "" {
		where += fmt.Sprintf(" AND (sku ILIKE $%d OR name ILIKE $%d)", n, n)
		args = append(args, "%"+search+"%")
		n++
	}

	var total int
	countQ := "SELECT COUNT(*) FROM products " + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf(`
		SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
			(SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id) AS last_movement_at
		FROM products p
		%s
		ORDER BY p.id
		LIMIT $%d OFFSET $%d
	`, where, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &p.LastMovementAt); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.Product{}
	}
	return out, total, rows.Err()
}

func (s *Store) GetProduct(ctx context.Context, id int) (*model.Product, error) {
	var p model.Product
	err := s.pool.QueryRow(ctx, `
		SELECT p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
			(SELECT MAX(m.created_at) FROM stock_movements m WHERE m.product_id = p.id)
		FROM products p WHERE p.id = $1
	`, id).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt, &p.LastMovementAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) CreateProduct(ctx context.Context, req model.CreateProductReq) (*model.Product, error) {
	stock := 0
	if req.Stock != nil {
		stock = *req.Stock
	}
	var p model.Product
	err := s.pool.QueryRow(ctx, `
		INSERT INTO products (sku, name, price, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, sku, name, price, stock, is_active, created_at
	`, strings.TrimSpace(req.SKU), strings.TrimSpace(req.Name), req.Price, stock).
		Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateSKU
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdateProduct(ctx context.Context, id int, req model.UpdateProductReq) (*model.Product, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	var p model.Product
	err := s.pool.QueryRow(ctx, `
		UPDATE products
		SET sku = $2, name = $3, price = $4, is_active = $5
		WHERE id = $1
		RETURNING id, sku, name, price, stock, is_active, created_at
	`, id, strings.TrimSpace(req.SKU), strings.TrimSpace(req.Name), req.Price, isActive).
		Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateSKU
		}
		return nil, err
	}
	_ = s.pool.QueryRow(ctx, `SELECT MAX(created_at) FROM stock_movements WHERE product_id = $1`, id).Scan(&p.LastMovementAt)
	return &p, nil
}

func (s *Store) DeleteProduct(ctx context.Context, id int) error {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	var hasHist bool
	err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM stock_movements WHERE product_id = $1)`, id).Scan(&hasHist)
	if err != nil {
		return err
	}
	if hasHist {
		return ErrProductHasHistory
	}

	ct, err := s.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AdjustStock(ctx context.Context, productID int, typ string, qty int, note string) (*model.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var p model.Product
	err = tx.QueryRow(ctx, `
		SELECT id, sku, name, price, stock, is_active, created_at
		FROM products WHERE id = $1 FOR UPDATE
	`, productID).Scan(&p.ID, &p.SKU, &p.Name, &p.Price, &p.Stock, &p.IsActive, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if typ == "out" {
		ct, err := tx.Exec(ctx, `
			UPDATE products SET stock = stock - $2
			WHERE id = $1 AND stock >= $2
		`, productID, qty)
		if err != nil {
			return nil, err
		}
		if ct.RowsAffected() == 0 {
			return nil, &InsufficientStockError{SKU: p.SKU, Available: p.Stock}
		}
		p.Stock -= qty
	} else {
		_, err = tx.Exec(ctx, `UPDATE products SET stock = stock + $2 WHERE id = $1`, productID, qty)
		if err != nil {
			return nil, err
		}
		p.Stock += qty
	}

	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (product_id, type, qty, note)
		VALUES ($1, $2, $3, $4)
	`, productID, typ, qty, notePtr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	p.LastMovementAt = &now

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) ListMovements(ctx context.Context, productID int) ([]model.StockMovement, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`, productID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, product_id, type, qty, note, created_at
		FROM stock_movements
		WHERE product_id = $1
		ORDER BY created_at DESC, id DESC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.StockMovement
	for rows.Next() {
		var m model.StockMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.Type, &m.Qty, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []model.StockMovement{}
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
