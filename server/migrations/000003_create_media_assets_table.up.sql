-- 媒体资源元数据：上传成功后落库，业务表通过 UUID 引用 media_assets.id
BEGIN;

CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size >= 0),
    asset_type TEXT NOT NULL DEFAULT 'image' CHECK (asset_type IN ('image', 'video')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE media_assets IS '对象存储媒体元数据：storage_key 对应桶内对象键';
COMMENT ON COLUMN media_assets.id IS '主键：媒体 UUID，供 vehicles.cover_media_id 等引用';
COMMENT ON COLUMN media_assets.storage_key IS '桶内对象键（路径+文件名），与 MinIO PutObject 一致';
COMMENT ON COLUMN media_assets.mime_type IS 'MIME 类型';
COMMENT ON COLUMN media_assets.file_size IS '字节大小';
COMMENT ON COLUMN media_assets.asset_type IS '媒体大类：image / video';
COMMENT ON COLUMN media_assets.created_at IS '创建时间';

CREATE UNIQUE INDEX idx_media_assets_storage_key ON media_assets (storage_key);

COMMIT;
