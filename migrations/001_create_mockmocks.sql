CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE mockmocks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    status      INT  NOT NULL DEFAULT 200,
    response    JSONB NOT NULL DEFAULT 'null',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (method, path)
);
