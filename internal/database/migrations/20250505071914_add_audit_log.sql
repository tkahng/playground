-- migrate:up
-- Step 1: Create ENUM type for log levels
-- Step 2: Create logs table
CREATE TABLE IF NOT EXISTS public.logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    level int NOT NULL DEFAULT 0,
    source text,
    message text NOT NULL,
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- Step 3: Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON public.logs (created_at);
CREATE INDEX IF NOT EXISTS idx_logs_level ON public.logs (level);
CREATE INDEX IF NOT EXISTS idx_logs_source ON public.logs (source);
CREATE INDEX IF NOT EXISTS idx_logs_data_gin ON public.logs USING GIN (data);
-- migrate:down
DROP INDEX IF EXISTS idx_logs_data_gin;
DROP INDEX IF EXISTS idx_logs_source;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_created_at;
DROP TABLE IF EXISTS public.logs;
DROP TYPE IF EXISTS log_level;