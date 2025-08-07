-- Database schema for ToDo List application
-- Updated to match the current database structure

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    display_name CHARACTER VARYING,
    email CHARACTER VARYING,
    location CHARACTER VARYING,
    bio TEXT,
    CONSTRAINT users_pkey PRIMARY KEY (id)
);

-- Tasks table (matches current structure)
CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id BIGINT NOT NULL,
    task_name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    priority SMALLINT NOT NULL DEFAULT 0,
    due_date TIMESTAMP WITH TIME ZONE,
    is_completed BOOLEAN NOT NULL DEFAULT false,
    is_recurring BOOLEAN NOT NULL DEFAULT false,
    recurring_frequency TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT tasks_pkey PRIMARY KEY (id),
    CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Habits table
CREATE TABLE IF NOT EXISTS habits (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    target_value TEXT,
    is_achieved BOOLEAN NOT NULL DEFAULT false,
    last_tracked_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT habits_pkey PRIMARY KEY (id),
    CONSTRAINT habits_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Logs table
CREATE TABLE IF NOT EXISTS logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id BIGINT,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT logs_pkey PRIMARY KEY (id),
    CONSTRAINT logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Pomodoro Sessions table
CREATE SEQUENCE IF NOT EXISTS pomodoro_sessions_id_seq;
CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id BIGINT NOT NULL DEFAULT nextval('pomodoro_sessions_id_seq'::regclass),
    user_id BIGINT NOT NULL,
    task_id BIGINT,
    duration INTEGER NOT NULL CHECK (duration > 0 AND duration <= 60),
    is_completed BOOLEAN DEFAULT false,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT pomodoro_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT pomodoro_sessions_task_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    CONSTRAINT pomodoro_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Goals table (additional table for enhanced functionality)
CREATE TABLE IF NOT EXISTS goals (
    id BIGINT GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL CHECK (category IN ('short-term', 'long-term')),
    target_value INTEGER NOT NULL CHECK (target_value > 0),
    current_value INTEGER DEFAULT 0 CHECK (current_value >= 0),
    unit TEXT NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    is_completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT goals_pkey PRIMARY KEY (id),
    CONSTRAINT goals_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_category ON tasks(category);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_is_completed ON tasks(is_completed);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);

CREATE INDEX IF NOT EXISTS idx_habits_user_id ON habits(user_id);
CREATE INDEX IF NOT EXISTS idx_habits_type ON habits(type);

CREATE INDEX IF NOT EXISTS idx_logs_user_id ON logs(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_event_type ON logs(event_type);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at);

CREATE INDEX IF NOT EXISTS idx_goals_user_id ON goals(user_id);
CREATE INDEX IF NOT EXISTS idx_goals_category ON goals(category);

CREATE INDEX IF NOT EXISTS idx_pomodoro_user_id ON pomodoro_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_pomodoro_task_id ON pomodoro_sessions(task_id);

-- Insert sample data (optional)
-- Insert a default admin user (password: admin)
INSERT INTO users (username, password_hash, display_name, email) 
VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Administrator', 'admin@todolist.com')
ON CONFLICT (username) DO NOTHING;

-- Get the admin user ID for sample tasks
DO $$
DECLARE
    admin_id BIGINT;
BEGIN
    SELECT id INTO admin_id FROM users WHERE username = 'admin';
    
    IF admin_id IS NOT NULL THEN
        -- Insert sample tasks with categories
        INSERT INTO tasks (user_id, task_name, description, category, priority, due_date, is_completed) VALUES
        (admin_id, 'Complete project documentation', 'Write comprehensive documentation for the new project', 'Work', 3, CURRENT_DATE + INTERVAL '1 day', false),
        (admin_id, 'Review code changes', 'Review and approve pending pull requests', 'Work', 2, CURRENT_DATE, false),
        (admin_id, 'Buy groceries', 'Buy milk, bread, and vegetables', 'Personal', 1, NULL, true),
        (admin_id, 'Schedule dentist appointment', 'Call dentist office for annual checkup', 'Health', 2, CURRENT_DATE + INTERVAL '3 days', false),
        (admin_id, 'Update portfolio website', 'Add recent projects to portfolio', 'Work', 2, CURRENT_DATE + INTERVAL '1 week', false),
        (admin_id, 'Exercise routine', 'Complete 30-minute workout', 'Health', 3, CURRENT_DATE, false),
        (admin_id, 'Read technical book', 'Read next chapter of Go programming book', 'Education', 1, NULL, false),
        (admin_id, 'Plan weekend trip', 'Research and book weekend getaway', 'Travel', 1, CURRENT_DATE + INTERVAL '2 weeks', false),
        (admin_id, 'Pay utility bills', 'Pay electricity and water bills', 'Finance', 3, CURRENT_DATE + INTERVAL '2 days', false),
        (admin_id, 'Learn new programming language', 'Start learning Rust programming', 'Education', 2, NULL, false)
        ON CONFLICT DO NOTHING;
        
        -- Insert sample habits
        INSERT INTO habits (user_id, name, type, target_value) VALUES
        (admin_id, 'Daily Exercise', 'health', '30 minutes'),
        (admin_id, 'Read Books', 'learning', '1 chapter'),
        (admin_id, 'Drink Water', 'health', '8 glasses'),
        (admin_id, 'Meditate', 'mindfulness', '10 minutes')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- Create a view for task statistics
CREATE OR REPLACE VIEW task_stats AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(t.id) as total_tasks,
    COUNT(CASE WHEN t.is_completed = true THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN t.is_completed = false THEN 1 END) as pending_tasks,
    COUNT(CASE WHEN t.due_date < CURRENT_DATE AND t.is_completed = false THEN 1 END) as overdue_tasks
FROM users u
LEFT JOIN tasks t ON u.id = t.user_id
GROUP BY u.id, u.username;
