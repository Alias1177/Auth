CREATE TABLE users (
    id SERIAL PRIMARY KEY ,
    email TEXT  unique NOT NULL ,
    password Text NOT NULL
)
