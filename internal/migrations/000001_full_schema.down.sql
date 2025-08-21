DROP INDEX IF EXISTS idx_deliveries_order_uid;
DROP INDEX IF EXISTS idx_payments_order_uid;
DROP INDEX IF EXISTS idx_products_track_number;
DROP INDEX IF EXISTS idx_products_track_number_chrt_id;

DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
