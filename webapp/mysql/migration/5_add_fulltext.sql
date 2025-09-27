-- backend/migrations/20250928_03_add_fulltext_and_sort_indexes.sql

-- 検索用（必須）
ALTER TABLE products
  ADD FULLTEXT INDEX ft_products_name_desc (name, description);

-- 任意: ソート最適化（LIMIT/OFFSET時の filesort 回避に効く）
CREATE INDEX idx_products_name_id   ON products (name,   product_id);
CREATE INDEX idx_products_value_id  ON products (value,  product_id);
CREATE INDEX idx_products_weight_id ON products (weight, product_id);
