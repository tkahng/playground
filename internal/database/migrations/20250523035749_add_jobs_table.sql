-- migrate:up
create type public.job_status AS ENUM ('pending', 'processing', 'done', 'failed');
CREATE TABLE public.jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind TEXT NOT NULL,
    unique_key TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status public.job_status NOT NULL DEFAULT 'pending',
    run_after TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX uniq_jobs_active_key ON public.jobs (unique_key)
WHERE status IN ('pending', 'processing');
CREATE INDEX jobs_polling_idx ON public.jobs (status, run_after, attempts);
-- migrate:down
DROP INDEX IF EXISTS jobs_polling_idx;
DROP INDEX IF EXISTS uniq_jobs_active_key;
DROP TABLE IF EXISTS public.jobs;
DROP TYPE IF EXISTS public.job_status;