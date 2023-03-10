// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	CreateApplication(ctx context.Context, arg CreateApplicationParams) (Application, error)
	//
	CreateLoginToken(ctx context.Context, arg CreateLoginTokenParams) (LoginToken, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	//
	CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error)
	DeleteLoginToken(ctx context.Context, token string) error
	DeleteUserSession(ctx context.Context, token string) error
	GetApplication(ctx context.Context, name string) (Application, error)
	GetLoginToken(ctx context.Context, token string) (LoginToken, error)
	GetUser(ctx context.Context, id int64) (User, error)
	//
	GetUserApplications(ctx context.Context, userID sql.NullInt64) ([]Application, error)
	GetUserByEmailOrUsername(ctx context.Context, arg GetUserByEmailOrUsernameParams) (User, error)
	GetUserSession(ctx context.Context, token string) (UserSession, error)
	UpdateApplication(ctx context.Context, arg UpdateApplicationParams) (Application, error)
}

var _ Querier = (*Queries)(nil)
