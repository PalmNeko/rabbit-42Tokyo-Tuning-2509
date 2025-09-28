-- Optimized performance indexes for products table
-- Based on actual query patterns in product.go

-- Essential composite indexes for common query patterns
-- LIKE search with sorting (covers the main ListProducts query)
CREATE INDEX idx_products_search_sort ON products(name, value, product_id);
CREATE INDEX idx_products_desc_search_sort ON products(description(50), value, product_id);

-- Individual sort indexes for different sort options
CREATE INDEX idx_products_value_sort ON products(value, product_id);
CREATE INDEX idx_products_weight_sort ON products(weight, product_id);
CREATE INDEX idx_products_name_sort ON products(name, product_id);

-- Simple index for name LIKE queries (fallback)
CREATE INDEX idx_products_name_prefix ON products(name);