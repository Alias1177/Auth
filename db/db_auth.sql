CREATE TABLE UsersLog (
    id SERIAL PRIMARY KEY ,
    username TEXT  UNIQUE NOT NULL,
    email TEXT  unique NOT NULL ,
    password Text NOT NULL
)
