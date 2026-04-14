-- 车型详情图与参数值：支撑发布校验、公开详情页参数表。
BEGIN;

CREATE TABLE vehicle_detail_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_assets (id) ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_vehicle_detail_media UNIQUE (vehicle_id, media_id)
);

COMMENT ON TABLE vehicle_detail_media IS '车型详情图集：车辆详情页轮播/大图';
COMMENT ON COLUMN vehicle_detail_media.vehicle_id IS '车型 ID';
COMMENT ON COLUMN vehicle_detail_media.media_id IS '媒体 ID';
COMMENT ON COLUMN vehicle_detail_media.sort_order IS '排序权重，越大越靠前';

CREATE INDEX idx_vehicle_detail_media_vehicle_sort ON vehicle_detail_media (vehicle_id, sort_order DESC, updated_at DESC);

CREATE TABLE vehicle_param_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles (id) ON DELETE CASCADE,
    template_item_id UUID NOT NULL REFERENCES param_template_items (id) ON DELETE RESTRICT,
    value_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_vehicle_param_values UNIQUE (vehicle_id, template_item_id)
);

COMMENT ON TABLE vehicle_param_values IS '车型参数值：按参数模板项录入';
COMMENT ON COLUMN vehicle_param_values.vehicle_id IS '车型 ID';
COMMENT ON COLUMN vehicle_param_values.template_item_id IS '参数模板项 ID';
COMMENT ON COLUMN vehicle_param_values.value_text IS '参数值文本，空字符串视为未填写';

CREATE INDEX idx_vehicle_param_values_vehicle ON vehicle_param_values (vehicle_id);
CREATE INDEX idx_vehicle_param_values_template_item ON vehicle_param_values (template_item_id);

COMMIT;
