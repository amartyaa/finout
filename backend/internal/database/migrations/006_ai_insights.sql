CREATE TABLE anomalies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    service TEXT NOT NULL,
    expected_amount DECIMAL(12,4),
    actual_amount DECIMAL(12,4),
    deviation_pct DECIMAL(8,2),
    confidence_score DECIMAL(5,2),
    narrative TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'dismissed')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE forecasts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    forecast_date DATE NOT NULL,
    predicted_total DECIMAL(12,4),
    best_case DECIMAL(12,4),
    worst_case DECIMAL(12,4),
    accuracy_pct DECIMAL(5,2),
    narrative TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    category TEXT NOT NULL CHECK (category IN ('rightsizing', 'idle_resource', 'reserved_instance')),
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    title TEXT NOT NULL,
    description TEXT,
    estimated_monthly_savings DECIMAL(12,4),
    risk_level TEXT NOT NULL DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high')),
    confidence_score DECIMAL(5,2),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'accepted', 'dismissed')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_anomalies_org ON anomalies(org_id, date);
CREATE INDEX idx_forecasts_org ON forecasts(org_id, forecast_date);
CREATE INDEX idx_recommendations_org ON recommendations(org_id, status);
