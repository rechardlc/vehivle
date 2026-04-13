-- 参数模板：与 doc/tech.md 2.1 一致；一级分类绑定一条模板，其下为参数项（文本/数值/单选）
BEGIN;

CREATE TABLE param_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 模板名称（后台展示）
    name TEXT NOT NULL,
    -- 绑定的一级分类（categories.level=1；约束由应用层校验）
    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    -- 启用 / 禁用：1=启用 0=禁用
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE param_templates IS '参数模板：按一级分类绑定（每分类至多一条）';
COMMENT ON COLUMN param_templates.id IS '主键：模板 UUID';
COMMENT ON COLUMN param_templates.name IS '模板名称';
COMMENT ON COLUMN param_templates.category_id IS '一级分类 ID（外键 categories）';
COMMENT ON COLUMN param_templates.status IS '状态：1=启用 0=禁用';
COMMENT ON COLUMN param_templates.created_at IS '创建时间';
COMMENT ON COLUMN param_templates.updated_at IS '最近更新时间';

CREATE UNIQUE INDEX uq_param_templates_category_id ON param_templates (category_id);
CREATE INDEX idx_param_templates_status ON param_templates (status);

CREATE TABLE param_template_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES param_templates (id) ON DELETE CASCADE,
    -- 机读键：同模板内唯一，用于 vehicle_param_values 等关联
    field_key VARCHAR(128) NOT NULL,
    -- 展示名称
    field_name TEXT NOT NULL,
    -- V1：text / number / single_select
    field_type TEXT NOT NULL CHECK (field_type IN ('text', 'number', 'single_select')),
    unit TEXT NULL,
    required SMALLINT NOT NULL DEFAULT 0 CHECK (required IN (0, 1)),
    -- 前台是否展示该参数项
    display SMALLINT NOT NULL DEFAULT 1 CHECK (display IN (0, 1)),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_param_template_items_template_field UNIQUE (template_id, field_key)
);

COMMENT ON TABLE param_template_items IS '参数模板项：字段定义（类型/单位/必填/展示/排序）';
COMMENT ON COLUMN param_template_items.id IS '主键：参数项 UUID';
COMMENT ON COLUMN param_template_items.template_id IS '所属模板 ID';
COMMENT ON COLUMN param_template_items.field_key IS '机读字段键，模板内唯一';
COMMENT ON COLUMN param_template_items.field_name IS '展示名称';
COMMENT ON COLUMN param_template_items.field_type IS '字段类型：text / number / single_select';
COMMENT ON COLUMN param_template_items.unit IS '单位（可选）';
COMMENT ON COLUMN param_template_items.required IS '是否必填：1=是 0=否';
COMMENT ON COLUMN param_template_items.display IS '前台是否展示：1=展示 0=隐藏';
COMMENT ON COLUMN param_template_items.sort_order IS '排序权重，越大越靠前';
COMMENT ON COLUMN param_template_items.created_at IS '创建时间';
COMMENT ON COLUMN param_template_items.updated_at IS '最近更新时间';

CREATE INDEX idx_param_template_items_template_sort ON param_template_items (template_id, sort_order DESC, updated_at DESC);

COMMIT;
