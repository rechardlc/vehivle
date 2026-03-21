-- 回滚 000001：按依赖顺序删除（当前无 FK，顺序可固定）
BEGIN;

DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS admin_users;

COMMIT;
