-- Add cloud_provider column to distinguish AWS vs Azure cost rows
ALTER TABLE daily_costs ADD COLUMN cloud_provider TEXT NOT NULL DEFAULT 'aws';

-- Drop old unique constraint and recreate with cloud_provider
ALTER TABLE daily_costs DROP CONSTRAINT IF EXISTS daily_costs_org_id_date_service_account_id_key;
ALTER TABLE daily_costs ADD CONSTRAINT daily_costs_org_date_service_account_provider_key
    UNIQUE(org_id, date, service, account_id, cloud_provider);

-- Index for filtering by provider
CREATE INDEX idx_daily_costs_provider ON daily_costs(org_id, cloud_provider);
