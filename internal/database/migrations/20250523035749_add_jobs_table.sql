-- migrate:up
create type public.job_status AS ENUM ('pending', 'processing', 'done', 'failed');
CREATE TABLE public.jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
    unique_key TEXT,
    payload JSONB,
    status public.job_status NOT NULL DEFAULT 'pending',
    run_after TIMESTAMPTZ DEFAULT now(),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    last_error TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE TRIGGER handle_jobs_updated_at BEFORE
UPDATE ON public.jobs FOR EACH ROW EXECUTE PROCEDURE set_current_timestamp_updated_at();
CREATE UNIQUE INDEX uniq_jobs_active_key ON public.jobs (unique_key)
WHERE status IN ('pending', 'processing');
CREATE INDEX idx_jobs_status_run_after ON public.jobs (status, run_after);
-- migrate:down
DROP INDEX IF EXISTS idx_jobs_status_run_after;
DROP INDEX IF EXISTS uniq_jobs_active_key;
DROP TRIGGER IF EXISTS handle_jobs_updated_at ON public.jobs;
DROP TABLE IF EXISTS public.jobs;
DROP TYPE IF EXISTS public.job_status;