-- migrate:up
create type app.job_status AS ENUM ('pending', 'processing', 'done', 'failed');
CREATE TABLE app.jobs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    kind TEXT NOT NULL,
    unique_key TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status app.job_status NOT NULL DEFAULT 'pending',
    run_after timestamptz not null default clock_timestamp(),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX uniq_jobs_active_key ON app.jobs (unique_key)
WHERE status IN ('pending', 'processing');
CREATE INDEX jobs_polling_idx ON app.jobs (status, run_after, attempts);
-- migrate:down
DROP INDEX IF EXISTS jobs_polling_idx;
DROP INDEX IF EXISTS uniq_jobs_active_key;
DROP TABLE IF EXISTS app.jobs;
DROP TYPE IF EXISTS app.job_status;