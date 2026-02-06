CREATE TABLE fare_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country_code VARCHAR(2) NOT NULL REFERENCES countries(code),
    service_type VARCHAR(20) NOT NULL REFERENCES service_types(id), 
    
    base_fare NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    cost_per_km NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    cost_per_minute NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    minimum_fare NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    
    is_active BOOLEAN NOT NULL DEFAULT true,
    version INT NOT NULL DEFAULT 1,
    version_by VARCHAR(50) NOT NULL DEFAULT 'system'
);

-- 1. Ensure uniqueness for National rates (where region is NULL)
CREATE UNIQUE INDEX uq_fare_rates_national 
ON fare_rates (country_code, service_type);

-- 3. lookup index
CREATE INDEX idx_fare_rates_lookup 
ON fare_rates (country_code, is_active);


CREATE TABLE fare_rates_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fare_rate_id UUID NOT NULL, 
    version INT NOT NULL,       -- 1 = Created, 2+ = Updated
    
    -- Snapshot of the data at that version
    base_fare NUMERIC(12, 2) NOT NULL,
    cost_per_km NUMERIC(12, 2) NOT NULL,
    cost_per_minute NUMERIC(12, 2) NOT NULL,
    minimum_fare NUMERIC(12, 2) NOT NULL,
    version_by VARCHAR(50) NOT NULL DEFAULT 'system',
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(fare_rate_id, version)
);

CREATE INDEX idx_audit_fare_id ON fare_rates_audit(fare_rate_id, version);

CREATE OR REPLACE FUNCTION fn_audit_fare_rates()
RETURNS TRIGGER AS $$
BEGIN
    -- If it's an UPDATE, increment the version in the main table record
    IF (TG_OP = 'UPDATE') THEN
        NEW.current_version = OLD.current_version + 1;
    END IF;

    -- Insert the snapshot into the audit table
    INSERT INTO fare_rates_audit (
        fare_rate_id,
        version,
        base_fare,
        cost_per_km,
        cost_per_minute,
        minimum_fare,
        is_active,
        version_by
    ) VALUES (
        NEW.id,
        NEW.current_version,
        NEW.base_fare,
        NEW.cost_per_km,
        NEW.cost_per_minute,
        NEW.minimum_fare,
        NEW.is_active,
        NEW.version_by
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trg_fare_rates_audit
AFTER INSERT OR UPDATE ON fare_rates
FOR EACH ROW
EXECUTE FUNCTION fn_audit_fare_rates();