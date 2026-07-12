CREATE TABLE routing_rules (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    url_id BIGINT NOT NULL,
    type VARCHAR(255) NOT NULL,
    config JSONB NOT NULL
);

CREATE INDEX idx_routing_rules_url_id ON routing_rules(url_id);
CREATE INDEX idx_routing_rules_deleted_at ON routing_rules(deleted_at);
