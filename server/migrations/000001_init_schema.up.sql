-- 首版 schema：与循序渐进「先三张主表」一致，后续用新编号迁移追加表/字段。
-- 字段语义对齐 PRD/领域：车型状态 draft/published/unpublished/deleted；价格模式 show_price/phone_inquiry。
BEGIN;

-- 后台账号：登录与简化 RBAC（超级管理员 / 运营编辑等）
CREATE TABLE admin_users (
    -- 主键：后台用户唯一标识（UUID）
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 登录名，全局唯一
    username VARCHAR(64) NOT NULL UNIQUE,
    -- 密码哈希（明文不落库）
    password_hash TEXT NOT NULL,
    -- 角色标识，如 super_admin / editor（与业务 RBAC 约定一致）
    role VARCHAR(32) NOT NULL DEFAULT 'editor',
    -- 创建时间（带时区）
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 最近更新时间（带时区）
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE admin_users IS '后台用户表：账号与角色';
COMMENT ON COLUMN admin_users.id IS '主键：后台用户 UUID';
COMMENT ON COLUMN admin_users.username IS '登录用户名，唯一';
COMMENT ON COLUMN admin_users.password_hash IS '密码哈希';
COMMENT ON COLUMN admin_users.role IS '角色：权限边界（如超级管理员、运营编辑）';
COMMENT ON COLUMN admin_users.created_at IS '创建时间';
COMMENT ON COLUMN admin_users.updated_at IS '最近更新时间';

-- 车型主数据：草稿/上架/下架/删除与展示字段
CREATE TABLE vehicles (
    -- 主键：车型唯一标识（UUID）；对外 API 可序列化为 string
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 所属分类 ID（两级分类落地后关联 categories；首版可空）
    category_id UUID NULL,
    -- 车型名称（展示用）
    name TEXT NOT NULL,
    -- 封面媒体资源 ID（关联 media_assets 等，首版可空）
    cover_media_id TEXT NULL,
    -- 价格展示策略：show_price=显示建议零售价；phone_inquiry=电话询价
    price_mode TEXT NOT NULL DEFAULT 'phone_inquiry' CHECK (price_mode IN ('show_price', 'phone_inquiry')),
    -- 建议零售价（单位与产品约定一致，未展示询价时仍可存 0）
    msrp_price INTEGER NOT NULL DEFAULT 0,
    -- 生命周期状态：draft/published/unpublished/deleted（仅 published 对小程序公开）
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'unpublished', 'deleted')),
    -- 卖点等富文本或纯文本（后续可接 XSS 过滤）
    selling_points TEXT NOT NULL DEFAULT '',
    -- 排序权重：数值越大越靠前（同值再按 updated_at）
    sort_order INTEGER NOT NULL DEFAULT 0,
    -- 创建时间
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 最近更新时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE vehicles IS '车型主表：内容、状态与排序';
COMMENT ON COLUMN vehicles.id IS '主键：车型 UUID';
COMMENT ON COLUMN vehicles.category_id IS '分类外键（预留，可空）';
COMMENT ON COLUMN vehicles.name IS '车型名称';
COMMENT ON COLUMN vehicles.cover_media_id IS '封面媒体 ID（预留）';
COMMENT ON COLUMN vehicles.price_mode IS '价格模式：展示零售价或电话询价';
COMMENT ON COLUMN vehicles.msrp_price IS '建议零售价（整数，单位与产品约定一致）';
COMMENT ON COLUMN vehicles.status IS '状态：草稿/已上架/已下架/逻辑删除';
COMMENT ON COLUMN vehicles.selling_points IS '卖点文案';
COMMENT ON COLUMN vehicles.sort_order IS '排序权重，越大越靠前';
COMMENT ON COLUMN vehicles.created_at IS '创建时间';
COMMENT ON COLUMN vehicles.updated_at IS '最近更新时间';

CREATE INDEX idx_vehicles_status ON vehicles (status);
CREATE INDEX idx_vehicles_category_id ON vehicles (category_id);

-- 全局键值配置：公司信息、联系方式、默认分享图、价格策略、免责声明等（JSON 承载灵活字段）
CREATE TABLE system_settings (
    -- 配置项键，全局唯一（如 contact_phone、disclaimer_text）
    setting_key VARCHAR(128) PRIMARY KEY,
    -- 配置值 JSON：结构随业务演进，读写由应用层约定
    setting_value JSONB NOT NULL DEFAULT '{}',
    -- 最近更新时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE system_settings IS '系统设置：键值存储（JSON）';
COMMENT ON COLUMN system_settings.setting_key IS '配置键，主键';
COMMENT ON COLUMN system_settings.setting_value IS '配置值（JSONB）';
COMMENT ON COLUMN system_settings.updated_at IS '最近更新时间';

COMMIT;
