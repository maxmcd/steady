CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL
);
CREATE TABLE services (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    user_id INTEGER NOT NULL
);
CREATE UNIQUE INDEX services_user_id_name_idx ON services(user_id, name);
CREATE TABLE service_versions (
    id INTEGER PRIMARY KEY,
    service_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    source TEXT NOT NULL
);
CREATE UNIQUE INDEX service_versions_service_id_version ON service_versions(service_id, version);
CREATE TABLE applications (
    id INTEGER PRIMARY KEY,
    service_version_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL
);
