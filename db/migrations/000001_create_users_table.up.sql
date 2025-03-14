CREATE TABLE IF NOT EXISTS userslog (
                                        id SERIAL PRIMARY KEY,
                                        username VARCHAR(255) NOT NULL,
                                        email VARCHAR(255) UNIQUE NOT NULL ,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

CREATE INDEX idx_userslog_email ON userslog(email);
CREATE INDEX idx_userslog_username ON userslog(username);

