-- backend/migrations/20250928_02_triggers_products_counter.sql
DROP TRIGGER IF EXISTS products_ai;
DROP TRIGGER IF EXISTS products_ad;

DELIMITER //

CREATE TRIGGER products_ai
AFTER INSERT ON products
FOR EACH ROW
BEGIN
  UPDATE product_counters SET total = total + 1 WHERE id = 1;
END//

CREATE TRIGGER products_ad
AFTER DELETE ON products
FOR EACH ROW
BEGIN
  UPDATE product_counters SET total = total - 1 WHERE id = 1;
END//

DELIMITER ;
