INSERT INTO products (sku, name, price, stock, is_active) VALUES
('SKU-BRG-001', 'Beras Premium 5kg', 78500, 120, true),
('SKU-BRG-002', 'Minyak Goreng 2L', 42000, 85, true),
('SKU-BRG-003', 'Gula Pasir 1kg', 17500, 200, true),
('SKU-BRG-004', 'Tepung Terigu 1kg', 14000, 150, true),
('SKU-BRG-005', 'Kopi Bubuk 200g', 28000, 45, true),
('SKU-BRG-006', 'Teh Celup 25s', 12500, 8, true),
('SKU-BRG-007', 'Susu UHT 1L', 18500, 60, true),
('SKU-BRG-008', 'Indomie Goreng (dus)', 115000, 30, true),
('SKU-BRG-009', 'Sabun Cuci Piring 800ml', 22000, 5, true),
('SKU-BRG-010', 'Deterjen Bubuk 1kg', 26500, 40, true),
('SKU-BRG-011', 'Tissue Gulung 12s', 32000, 0, false),
('SKU-BRG-012', 'Air Mineral 600ml (dus)', 48000, 95, true),
('SKU-BRG-013', 'Kecap Manis 600ml', 19500, 3, true),
('SKU-BRG-014', 'Sambal Botol 340ml', 16000, 70, true);

-- initial stock movements so last_movement_at is not null for most products
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'in', stock, 'Initial stock', now() - (random() * interval '10 days')
FROM products
WHERE stock > 0;

INSERT INTO sales_orders (customer, status, total, created_at) VALUES
('Toko Maju Jaya', 'CREATED', 156500, now() - interval '2 days'),
('Warung Bu Siti', 'CREATED', 84000, now() - interval '1 day');

INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES
(1, 'SKU-BRG-001', 1, 78500),
(1, 'SKU-BRG-002', 1, 42000),
(1, 'SKU-BRG-005', 1, 28000),
(1, 'SKU-BRG-006', 1, 12500),
(2, 'SKU-BRG-002', 2, 42000);

-- reflect those sample orders in stock + movements
UPDATE products SET stock = stock - 1 WHERE sku = 'SKU-BRG-001';
UPDATE products SET stock = stock - 3 WHERE sku = 'SKU-BRG-002';
UPDATE products SET stock = stock - 1 WHERE sku = 'SKU-BRG-005';
UPDATE products SET stock = stock - 1 WHERE sku = 'SKU-BRG-006';

INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'out', 1, 'SO #1', now() - interval '2 days' FROM products WHERE sku = 'SKU-BRG-001';
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'out', 1, 'SO #1', now() - interval '2 days' FROM products WHERE sku = 'SKU-BRG-002';
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'out', 1, 'SO #1', now() - interval '2 days' FROM products WHERE sku = 'SKU-BRG-005';
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'out', 1, 'SO #1', now() - interval '2 days' FROM products WHERE sku = 'SKU-BRG-006';
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'out', 2, 'SO #2', now() - interval '1 day' FROM products WHERE sku = 'SKU-BRG-002';
