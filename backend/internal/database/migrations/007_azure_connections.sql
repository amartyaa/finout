CREATE TABLE azure_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    subscription_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    last_sync_at TIMESTAMPTZ,
    error_message TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(org_id)
);
