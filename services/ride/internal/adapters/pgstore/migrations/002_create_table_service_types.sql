-- 1. Operational Table
CREATE TABLE service_types (
    id VARCHAR(20) PRIMARY KEY, -- 'standard', 'premium', 'xl'
    display_name VARCHAR(50) NOT NULL,
    max_passengers SMALLINT NOT NULL DEFAULT 4,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    current_version INT NOT NULL DEFAULT 1,
    version_by VARCHAR(50) NOT NULL DEFAULT 'system'
);

-- 2. Audit Table
CREATE TABLE service_types_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_type_id VARCHAR(20) NOT NULL REFERENCES service_types(id),
    version INT NOT NULL,
    
    display_name VARCHAR(50) NOT NULL,
    max_passengers SMALLINT NOT NULL,
    sort_order INT NOT NULL,
    is_active BOOLEAN NOT NULL,
    
    version_by VARCHAR(50) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(service_type_id, version)
);

CREATE INDEX idx_audit_service_type_id ON service_types_audit(service_type_id, version);

CREATE OR REPLACE FUNCTION fn_audit_service_types()
RETURNS TRIGGER AS $$
BEGIN
    -- Increment version on update in the main table
    IF (TG_OP = 'UPDATE') THEN
        NEW.current_version = OLD.current_version + 1;
    END IF;

    -- Log the snapshot with the version to the audit table
    INSERT INTO service_types_audit (
        service_type_id,
        version,
        display_name,
        max_passengers,
        sort_order,
        is_active,
        version_by
    ) VALUES (
        NEW.id,
        NEW.current_version,
        NEW.display_name,
        NEW.max_passengers,
        NEW.sort_order,
        NEW.is_active,
        NEW.version_by
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_service_types_audit
AFTER INSERT OR UPDATE ON service_types
FOR EACH ROW
EXECUTE FUNCTION fn_audit_service_types();