CREATE TABLE user_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(320) UNIQUE, -- Nullable: Social might not provide it, or phone-first signups won't have it initially
    phone           VARCHAR(20) UNIQUE,  -- Nullable: Pure social signups won't have this until prompted
    password_hash   VARCHAR(255) NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'rider' 
        CHECK (role IN ('rider', 'driver', 'admin')),
    status          VARCHAR(20) NOT NULL DEFAULT 'active' 
        CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_identity_presence CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE INDEX idx_credentials_email ON user_credentials(email);
CREATE INDEX idx_credentials_phone ON user_credentials(phone);
CREATE INDEX idx_credentials_status ON user_credentials(status) WHERE status = 'active';

CREATE TABLE user_identities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES user_credentials(id) ON DELETE CASCADE,
    provider        VARCHAR(20) NOT NULL
        CHECK (provider IN ('google', 'apple', 'facebook', 'phone_otp')),
    provider_user_id VARCHAR(255) NOT NULL,
    raw_data        JSONB,
    linked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_identities_lookup ON user_identities(provider, provider_user_id);