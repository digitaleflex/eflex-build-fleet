ALTER TABLE build_jobs ADD COLUMN required_capabilities TEXT[] NOT NULL DEFAULT '{}';
CREATE INDEX build_jobs_required_capabilities_idx ON build_jobs USING GIN (required_capabilities);