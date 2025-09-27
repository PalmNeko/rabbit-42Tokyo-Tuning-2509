-- backend/migrations/20250928_01_create_product_counters.sql
CREATE TABLE IF NOT EXISTS product_counters (
  id TINYINT NOT NULL PRIMARY KEY,
  total BIGINT NOT NULL
) ENGINE=InnoDB;

INSERT IGNORE INTO product_counters (id, total)
SELECT 1, COUNT(*) FROM products;
