-- 回滚 000002：先删 vehicles 外键，再删 categories
BEGIN;

ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS fk_vehicles_category_id;

DROP TABLE IF EXISTS categories;

COMMIT;
