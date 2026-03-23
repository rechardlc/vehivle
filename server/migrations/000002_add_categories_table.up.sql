-- 新增 categories：两级分类（一级金刚区 / 二级筛选标签），对齐 PRD 分类管理与 domain model.Category
BEGIN;

CREATE TABLE categories (
    -- 主键：分类 UUID
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 父级分类；一级分类为 NULL
    parent_id UUID NULL REFERENCES categories (id) ON DELETE RESTRICT,
    -- 层级：1=一级，2=二级（与「左侧一级 + 右侧二级筛选」一致）
    level SMALLINT NOT NULL DEFAULT 1 CHECK (level IN (1, 2)),
    -- 展示名称
    name TEXT NOT NULL,
    -- 启用 / 禁用：1=启用 0=禁用（禁用后前台不展示，后台可保留数据）
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    -- 排序权重：越大越靠前；同值再按 updated_at
    sort_order INTEGER NOT NULL DEFAULT 0,
    -- 分类图标媒体 ID（预留，关联 media_assets）
    icon_media_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE categories IS '车型分类表：支持两级树与排序、状态';
COMMENT ON COLUMN categories.id IS '主键：分类 UUID';
COMMENT ON COLUMN categories.parent_id IS '父分类 ID，一级分类为空';
COMMENT ON COLUMN categories.level IS '层级：1 一级 / 2 二级';
COMMENT ON COLUMN categories.name IS '分类名称（展示）';
COMMENT ON COLUMN categories.status IS '状态：1=启用 0=禁用';
COMMENT ON COLUMN categories.sort_order IS '排序权重，越大越靠前';
COMMENT ON COLUMN categories.icon_media_id IS '图标媒体 ID（预留）';
COMMENT ON COLUMN categories.created_at IS '创建时间';
COMMENT ON COLUMN categories.updated_at IS '最近更新时间';

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
CREATE INDEX idx_categories_status ON categories (status);
CREATE INDEX idx_categories_level_sort ON categories (level, sort_order DESC, updated_at DESC);

-- 首版 vehicles.category_id 已预留；补全与 categories 的外键（无效 UUID 需先清洗为 NULL）
ALTER TABLE vehicles
    ADD CONSTRAINT fk_vehicles_category_id
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL;

COMMIT;
