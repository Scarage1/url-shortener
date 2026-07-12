CREATE TABLE plans (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    max_links INTEGER,
    max_redirects INTEGER,
    max_api_calls INTEGER,
    max_domains INTEGER,
    max_geo_rules INTEGER,
    max_password_links INTEGER,
    max_schedule_links INTEGER,
    max_members INTEGER,
    rate_limit INTEGER,
    price_monthly INTEGER,
    price_yearly INTEGER
);

CREATE UNIQUE INDEX idx_plans_name ON plans(name);
CREATE INDEX idx_plans_deleted_at ON plans(deleted_at);

CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    organization_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    status VARCHAR(255) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_subscriptions_organization_id ON subscriptions(organization_id);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions(deleted_at);
