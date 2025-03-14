CREATE TABLE IF NOT EXISTS userslog (
                                        id SERIAL PRIMARY KEY,
                                        username TEXT NOT NULL,
                                        email TEXT UNIQUE NOT NULL CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    password TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

CREATE INDEX idx_userslog_email ON userslog(email);
CREATE INDEX idx_userslog_username ON userslog(username);

-- Триггер для автоматического обновления updated_at при изменениях
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON userslog
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
