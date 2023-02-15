-- TODO: swap out all TEXT for appropriate VARCHAR
PRAGMA foreign_keys = ON;
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL
);
CREATE TABLE login_tokens (
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE user_sessions (
    user_id INTEGER NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE applications (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    name TEXT UNIQUE NOT NULL
);
