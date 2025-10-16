-- migrate:up
-- Step 2: Create logs table
CREATE TABLE IF NOT EXISTS app.audit_logs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    level int NOT NULL DEFAULT 0,
    source text,
    message text NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz not null default clock_timestamp()
);
-- Step 3: Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON app.audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_logs_level ON app.audit_logs (level);
CREATE INDEX IF NOT EXISTS idx_logs_source ON app.audit_logs (source);
CREATE INDEX IF NOT EXISTS idx_logs_data_gin ON app.audit_logs USING GIN (data);
-- migrate:down
DROP INDEX IF EXISTS idx_logs_data_gin;
DROP INDEX IF EXISTS idx_logs_source;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_created_at;
DROP TABLE IF EXISTS app.audit_logs;