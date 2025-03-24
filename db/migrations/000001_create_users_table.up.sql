CREATE TABLE IF NOT EXISTS UsersLog (
                                        id SERIAL PRIMARY KEY,
                                        username VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Добавьте проверку на существование индексов
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_userslog_email') THEN
CREATE INDEX idx_UsersLog_email ON UsersLog(email);
END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_userslog_username') THEN
CREATE INDEX idx_UsersLog_username ON UsersLog(username);
END IF;
END
$$;