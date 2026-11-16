CREATE TABLE node_credentials (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 node_id UUID NOT NULL REFERENCES build_nodes(id) ON DELETE CASCADE,
 token_hash BYTEA NOT NULL UNIQUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 last_used_at TIMESTAMPTZ,
 revoked_at TIMESTAMPTZ
);
CREATE INDEX node_credentials_active_idx ON node_credentials(node_id) WHERE revoked_at IS NULL;
