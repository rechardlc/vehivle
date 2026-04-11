BEGIN;

DROP TABLE IF EXISTS system_settings;

-- 恢复旧版 key-value 结构
CREATE TABLE system_settings (
    setting_key   VARCHAR(128) PRIMARY KEY,
    setting_value JSONB        NOT NULL DEFAULT '{}',
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

COMMIT;
