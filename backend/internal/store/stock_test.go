package store_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/hari/stockflow/internal/db"
	"github.com/hari/stockflow/internal/model"
	"github.com/hari/stockflow/internal/store"
)

func setup(t *testing.T) (*store.Store, func()) {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://stockflow:stockflow@localhost:5433/stockflow?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, url)
	if err != nil {
		t.Skipf("db unavailable: %v", err)
	}

	// isolate each test with a temp product
	st := store.New(pool)
	return st, func() { pool.Close() }
}

func TestAdjustStock_InsufficientRejectsNothingWritten(t *testing.T) {
	st, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	stock := 5
	p, err := st.CreateProduct(ctx, model.CreateProductReq{
		SKU:   "TEST-OUT-" + t.Name(),
		Name:  "Test Out",
		Price: 1000,
		Stock: &stock,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() {
		// cleanup movements then product
		_, _ = st.Pool().Exec(ctx, `DELETE FROM stock_movements WHERE product_id = $1`, p.ID)
		_, _ = st.Pool().Exec(ctx, `DELETE FROM products WHERE id = $1`, p.ID)
	}()

	_, err = st.AdjustStock(ctx, p.ID, "out", 99, "should fail")
	var stockErr *store.InsufficientStockError
	if !errors.As(err, &stockErr) {
		t.Fatalf("expected InsufficientStockError, got %v", err)
	}
	if stockErr.Available != 5 {
		t.Fatalf("available want 5 got %d", stockErr.Available)
	}

	got, err := st.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stock != 5 {
		t.Fatalf("stock mutated, want 5 got %d", got.Stock)
	}

	movs, err := st.ListMovements(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movs) != 0 {
		t.Fatalf("expected no movements, got %d", len(movs))
	}
}

func TestAdjustStock_HappyPath(t *testing.T) {
	st, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	stock := 10
	p, err := st.CreateProduct(ctx, model.CreateProductReq{
		SKU:   "TEST-IN-" + t.Name(),
		Name:  "Test In",
		Price: 2500,
		Stock: &stock,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer func() {
		_, _ = st.Pool().Exec(ctx, `DELETE FROM stock_movements WHERE product_id = $1`, p.ID)
		_, _ = st.Pool().Exec(ctx, `DELETE FROM products WHERE id = $1`, p.ID)
	}()

	updated, err := st.AdjustStock(ctx, p.ID, "out", 3, "picked")
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if updated.Stock != 7 {
		t.Fatalf("stock want 7 got %d", updated.Stock)
	}

	movs, err := st.ListMovements(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(movs) != 1 {
		t.Fatalf("want 1 movement, got %d", len(movs))
	}
	if movs[0].Type != "out" || movs[0].Qty != 3 {
		t.Fatalf("unexpected movement: %+v", movs[0])
	}
}

func TestCreateOrder_InsufficientRollsBack(t *testing.T) {
	st, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	a := 2
	b := 100
	pa, err := st.CreateProduct(ctx, model.CreateProductReq{SKU: "ORD-A-" + t.Name(), Name: "A", Price: 1000, Stock: &a})
	if err != nil {
		t.Fatal(err)
	}
	pb, err := st.CreateProduct(ctx, model.CreateProductReq{SKU: "ORD-B-" + t.Name(), Name: "B", Price: 2000, Stock: &b})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = st.Pool().Exec(ctx, `DELETE FROM sales_order_items WHERE sku IN ($1,$2)`, pa.SKU, pb.SKU)
		_, _ = st.Pool().Exec(ctx, `DELETE FROM stock_movements WHERE product_id IN ($1,$2)`, pa.ID, pb.ID)
		_, _ = st.Pool().Exec(ctx, `DELETE FROM products WHERE id IN ($1,$2)`, pa.ID, pb.ID)
	}()

	_, err = st.CreateOrder(ctx, "Test Cust", []model.OrderItemReq{
		{SKU: pa.SKU, Qty: 1},
		{SKU: pb.SKU, Qty: 999}, // blows up
	})
	var stockErr *store.InsufficientStockError
	if !errors.As(err, &stockErr) {
		t.Fatalf("want insufficient, got %v", err)
	}

	ga, _ := st.GetProduct(ctx, pa.ID)
	gb, _ := st.GetProduct(ctx, pb.ID)
	if ga.Stock != 2 || gb.Stock != 100 {
		t.Fatalf("stock leaked: a=%d b=%d", ga.Stock, gb.Stock)
	}
}
