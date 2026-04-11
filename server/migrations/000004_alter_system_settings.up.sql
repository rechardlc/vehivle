BEGIN;

-- 旧表是 key-value 结构（setting_key + setting_value JSONB），无法约束必填字段。
-- 重建为单行配置表：必填字段 NOT NULL，可选字段允许 NULL。
DROP TABLE IF EXISTS system_settings;

CREATE TABLE system_settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    company_name        TEXT        NOT NULL,
    customer_service_phone TEXT     NULL,
    customer_service_wechat TEXT    NULL,
    default_price_mode  TEXT        NOT NULL DEFAULT 'phone_inquiry'
        CHECK (default_price_mode IN ('show_price', 'phone_inquiry')),
    disclaimer_text     TEXT        NULL,
    default_share_title TEXT        NULL,
    default_share_image TEXT        NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE  system_settings IS '系统全局配置（单行）';
COMMENT ON COLUMN system_settings.company_name IS '公司名称（必填）';
COMMENT ON COLUMN system_settings.customer_service_phone IS '客服电话（未配置则前端隐藏）';
COMMENT ON COLUMN system_settings.customer_service_wechat IS '客服微信（未配置则前端隐藏）';
COMMENT ON COLUMN system_settings.default_price_mode IS '默认价格展示模式：show_price / phone_inquiry';
COMMENT ON COLUMN system_settings.disclaimer_text IS '免责声明文案';
COMMENT ON COLUMN system_settings.default_share_title IS '默认分享标题';
COMMENT ON COLUMN system_settings.default_share_image IS '默认分享图 storage_key';

COMMIT;
