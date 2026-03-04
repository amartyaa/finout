CREATE TABLE daily_costs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    service TEXT NOT NULL,
    account_id TEXT DEFAULT '',
    environment TEXT DEFAULT '',
    amount DECIMAL(12,4) NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id, date, service, account_id)
);

CREATE INDEX idx_daily_costs_org_date ON daily_costs(org_id, date);
CREATE INDEX idx_daily_costs_service ON daily_costs(org_id, service);
