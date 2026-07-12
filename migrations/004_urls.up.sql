CREATE TABLE urls (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    organization_id BIGINT NOT NULL,
    created_by BIGINT NOT NULL,
    short_code VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL,
    click_count INTEGER DEFAULT 0,
    last_accessed TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_organization_id ON urls(organization_id);
CREATE INDEX idx_urls_deleted_at ON urls(deleted_at);
