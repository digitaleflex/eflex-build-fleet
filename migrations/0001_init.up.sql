CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE build_nodes (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 node_name TEXT NOT NULL UNIQUE,
 status TEXT NOT NULL CHECK (status IN ('pending','ready','draining','offline','quarantined')),
 agent_version TEXT,
 last_heartbeat_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE node_capabilities (
 node_id UUID NOT NULL REFERENCES build_nodes(id) ON DELETE CASCADE,
 capability TEXT NOT NULL,
 value JSONB NOT NULL DEFAULT '{}'::jsonb,
 PRIMARY KEY (node_id, capability)
);
CREATE TABLE build_jobs (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 tenant_id TEXT NOT NULL,
 idempotency_key TEXT,
 status TEXT NOT NULL CHECK (status IN ('queued','scheduled','running','succeeded','failed','cancelled')),
 source_commit TEXT NOT NULL,
 engine TEXT,
 engine_version TEXT,
 requested_resources JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (tenant_id, idempotency_key)
);
CREATE INDEX build_jobs_queue_idx ON build_jobs (created_at) WHERE status = 'queued';
CREATE TABLE build_attempts (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 job_id UUID NOT NULL REFERENCES build_jobs(id) ON DELETE CASCADE,
 node_id UUID REFERENCES build_nodes(id),
 attempt_number INTEGER NOT NULL CHECK (attempt_number > 0),
 status TEXT NOT NULL CHECK (status IN ('assigned','running','succeeded','failed','lost')),
 started_at TIMESTAMPTZ, finished_at TIMESTAMPTZ, error_code TEXT,
 UNIQUE (job_id, attempt_number)
);
CREATE TABLE job_leases (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 job_id UUID NOT NULL REFERENCES build_jobs(id) ON DELETE CASCADE,
 attempt_id UUID NOT NULL REFERENCES build_attempts(id) ON DELETE CASCADE,
 node_id UUID NOT NULL REFERENCES build_nodes(id),
 fencing_token BIGINT NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (job_id, fencing_token)
);
CREATE INDEX job_leases_expiry_idx ON job_leases (expires_at);
CREATE TABLE job_events (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 job_id UUID NOT NULL REFERENCES build_jobs(id) ON DELETE CASCADE,
 event_type TEXT NOT NULL, payload JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE outbox_events (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 aggregate_type TEXT NOT NULL, aggregate_id UUID NOT NULL,
 event_type TEXT NOT NULL, payload JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(), published_at TIMESTAMPTZ
);
CREATE INDEX outbox_events_pending_idx ON outbox_events (id) WHERE published_at IS NULL;
CREATE TABLE audit_events (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 actor_type TEXT NOT NULL, actor_id TEXT, action TEXT NOT NULL,
 resource_type TEXT NOT NULL, resource_id TEXT,
 metadata JSONB NOT NULL DEFAULT '{}'::jsonb, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
