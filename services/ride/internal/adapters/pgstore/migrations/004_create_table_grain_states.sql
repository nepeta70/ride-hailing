CREATE TABLE grain_events (
    grain_kind   VARCHAR(50)  NOT NULL,
    grain_id     UUID         NOT NULL,
    version      INT          NOT NULL,
    event_type   VARCHAR(50)  NOT NULL,
    payload      JSONB        NOT NULL,
    timestamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (grain_kind, grain_id, version)
);

CREATE TABLE grain_snapshots (
    grain_kind   VARCHAR(50)  NOT NULL,
    grain_id     UUID         NOT NULL,
    version      INT          NOT NULL,
    core         JSONB        NOT NULL,
    state        JSONB        NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (grain_kind, grain_id)
);