package db

import _ "embed"

//go:embed migration.sql
var Migrations string

//go:generate sqlc generate
