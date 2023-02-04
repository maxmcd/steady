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
CREATE TABLE services (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE UNIQUE INDEX services_user_id_name_idx ON services(user_id, name);
CREATE TABLE service_versions (
    id INTEGER PRIMARY KEY,
    service_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    source TEXT NOT NULL,
    FOREIGN KEY (service_id) REFERENCES service (id)
);
CREATE UNIQUE INDEX service_versions_service_id_version ON service_versions(service_id, version);
CREATE TABLE applications (
    id INTEGER PRIMARY KEY,
    service_version_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    FOREIGN KEY (service_version_id) REFERENCES service_version (id)
);
