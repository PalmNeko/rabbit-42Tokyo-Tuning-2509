-- Performance optimization indexes for products table
-- This migration adds indexes to improve search and sort performance

-- Individual indexes for search fields
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_description ON products(description(128));

-- Indexes for common sort fields
CREATE INDEX idx_products_value ON products(value);
CREATE INDEX idx_products_weight ON products(weight);

-- Full-text search index for better Japanese text search
-- Note: Requires MySQL 5.7+ with ngram parser for Japanese
CREATE FULLTEXT INDEX idx_products_fulltext ON products(name, description) WITH PARSER ngram;

-- Composite indexes for common query patterns
-- Search by name with value sorting
CREATE INDEX idx_products_name_value ON products(name, value);

-- Search by description with value sorting
CREATE INDEX idx_products_description_value ON products(description(128), value);

-- General purpose composite index for pagination with sorting
CREATE INDEX idx_products_value_id ON products(value, product_id);
CREATE INDEX idx_products_weight_id ON products(weight, product_id);