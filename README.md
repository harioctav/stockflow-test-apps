# StockFlow

Small inventory + sales app for a warehouse supervisor: manage products, adjust stock, create orders that consume stock.

- Backend: Go + PostgreSQL
- Frontend: React + TypeScript (Vite)

## Run

### Full docker

```bash
docker compose up --build
```

- UI: http://localhost:5173
- API: http://localhost:8081
- DB: localhost:5433 (user/pass/db: `stockflow`)

Schema + seed are applied on first Postgres start via `backend/migrations/`.

Ports are 5433 / 8081 so they don't clash with a local Postgres or other apps on 5432/8080.

### Local dev

```bash
docker compose up -d db

cd backend && go run ./cmd/server

cd frontend && npm install && npm run dev
```

Vite proxies `/api` to `localhost:8081`.

Reset DB/seed:

```bash
docker compose down -v && docker compose up -d db
```

## Curl examples

```bash
curl 'http://localhost:8081/api/products?search=beras&page=1&limit=20'

curl -X POST http://localhost:8081/api/products \
  -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-TEST-1","name":"Sampel","price":15000,"stock":10}'

curl -X POST http://localhost:8081/api/products/1/stock \
  -H 'Content-Type: application/json' \
  -d '{"type":"out","qty":2,"note":"damaged"}'

curl -X POST http://localhost:8081/api/orders \
  -H 'Content-Type: application/json' \
  -d '{"customer":"Toko ABC","items":[{"sku":"SKU-BRG-001","qty":1},{"sku":"SKU-BRG-006","qty":2}]}'

curl http://localhost:8081/api/orders/1
curl -X POST http://localhost:8081/api/orders/1/cancel
```

Overdraw response:

```json
{"error":{"code":"INSUFFICIENT_STOCK","available":3,"sku":"SKU-BRG-006","message":"insufficient stock"}}
```

## Notes / decisions

- `last_movement_at` is computed from `stock_movements` (subquery), not stored on `products`.
- Stock never goes negative: `UPDATE ... SET stock = stock - n WHERE stock >= n`. Orders do this inside one transaction with `FOR UPDATE`.
- Summary cards are calculated on the frontend from the product list (fine for this dataset size).
- CORS is enabled; Docker frontend also proxies `/api` through nginx.

## Tests

```bash
cd backend
DATABASE_URL='postgres://stockflow:stockflow@localhost:5433/stockflow?sslmode=disable' go test ./internal/store/ -count=1

cd frontend
npm test
```

Backend covers: insufficient stock writes nothing, happy-path adjust, order rollback when one SKU is short, plus a concurrent adjust race check.

## If I had more time

- summary/report endpoint instead of loading all products for cards
- Idempotency-Key on create order
- server-side filters + real pagination UI
- soft-delete for products (SKU is referenced by order items)

## Time spent

Around 7 hours. Most of it on the stock/order transaction stuff; UI is kept simple on purpose.
