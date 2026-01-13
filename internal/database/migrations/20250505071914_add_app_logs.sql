-- migrate:up
-- Step 2: Create logs table
CREATE TABLE IF NOT EXISTS app.logs (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    level int NOT NULL DEFAULT 0,
    message text NOT NULL DEFAULT '',
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz not null default clock_timestamp()
);
-- Step 3: Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_logs_level ON app.logs (level);
CREATE INDEX IF NOT EXISTS idx_logs_message ON app.logs (message);
CREATE INDEX IF NOT EXISTS idx_logs_data_gin ON app.logs USING GIN (data);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON app.logs (created_at);
-- migrate:down
DROP INDEX IF EXISTS idx_logs_created_at;
DROP INDEX IF EXISTS idx_logs_data_gin;
DROP INDEX IF EXISTS idx_logs_message;
DROP INDEX IF EXISTS idx_logs_level;
DROP TABLE IF EXISTS app.logs;