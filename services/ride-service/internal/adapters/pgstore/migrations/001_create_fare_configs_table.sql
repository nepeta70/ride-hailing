CREATE TABLE fare_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_code VARCHAR(2) NOT NULL, -- e.g., 'es', 'us'
    service_type VARCHAR(20) NOT NULL, -- e.g., 'standard', 'premium', 'xl'
    currency_code VARCHAR(3) NOT NULL, -- ISO 4217 (EUR, USD)
    
    base_fare NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    cost_per_km NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    cost_per_minute NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    
    minimum_fare NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Ensure we only have one config per region/type
    UNIQUE(region_code, service_type)
);

-- Index for fast lookup when calculating estimates
CREATE INDEX idx_fare_lookup ON fare_configs (region_code, is_active);