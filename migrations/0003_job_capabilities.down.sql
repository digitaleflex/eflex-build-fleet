DROP INDEX IF EXISTS build_jobs_required_capabilities_idx;
ALTER TABLE build_jobs DROP COLUMN IF EXISTS required_capabilities;