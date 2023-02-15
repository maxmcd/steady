// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"database/sql"
	"time"
)

type Application struct {
	ID     int64
	UserID sql.NullInt64
	Name   string
}

type LoginToken struct {
	UserID    int64
	Token     string
	CreatedAt time.Time
}

type User struct {
	ID       int64
	Email    string
	Username string
}

type UserSession struct {
	UserID    int64
	Token     string
	CreatedAt time.Time
}
