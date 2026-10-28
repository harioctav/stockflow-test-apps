package store_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hari/stockflow/internal/model"
)

// spin concurrent outs against a small stock — stock must never go negative
func TestAdjustStock_ConcurrentNoNegative(t *testing.T) {
	st, cleanup := setup(t)
	defer cleanup()
	ctx := context.Background()

	stock := 20
	p, err := st.CreateProduct(ctx, model.CreateProductReq{
		SKU:   "RACE-" + t.Name(),
		Name:  "Race Product",
		Price: 1000,
		Stock: &stock,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = st.Pool().Exec(ctx, `DELETE FROM stock_movements WHERE product_id = $1`, p.ID)
		_, _ = st.Pool().Exec(ctx, `DELETE FROM products WHERE id = $1`, p.ID)
	}()

	var okCount atomic.Int32
	var wg sync.WaitGroup
	workers := 40
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := st.AdjustStock(ctx, p.ID, "out", 1, "race")
			if err == nil {
				okCount.Add(1)
			}
		}()
	}
	wg.Wait()

	got, err := st.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Stock < 0 {
		t.Fatalf("negative stock: %d", got.Stock)
	}
	if int(okCount.Load())+got.Stock != 20 {
		t.Fatalf("inconsistent: ok=%d remaining=%d", okCount.Load(), got.Stock)
	}
	if okCount.Load() != 20 {
		t.Fatalf("expected exactly 20 successful outs, got %d (stock=%d)", okCount.Load(), got.Stock)
	}
}
