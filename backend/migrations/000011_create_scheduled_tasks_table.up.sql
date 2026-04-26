-- Scheduled tasks table
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    prompt TEXT NOT NULL,
    schedule_type VARCHAR(20) NOT NULL DEFAULT 'daily' CHECK (schedule_type IN ('daily', 'weekly', 'cron')),
    schedule_time VARCHAR(50) NOT NULL DEFAULT '09:00',
    model VARCHAR(100) DEFAULT 'auto',
    enabled BOOLEAN DEFAULT true,
    notify_on_done BOOLEAN DEFAULT true,
    execution_count INT DEFAULT 0,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scheduled_tasks_user_id ON scheduled_tasks(user_id);
CREATE INDEX idx_scheduled_tasks_enabled ON scheduled_tasks(enabled);
