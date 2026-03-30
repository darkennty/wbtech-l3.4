CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    original_path TEXT NOT NULL,
    watermark_path TEXT,
    thumb_path TEXT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);