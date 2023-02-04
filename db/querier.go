// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"context"
)

type Querier interface {
	CreateService(ctx context.Context, arg CreateServiceParams) (Service, error)
	CreateServiceVersion(ctx context.Context, arg CreateServiceVersionParams) (ServiceVersion, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetService(ctx context.Context, arg GetServiceParams) (Service, error)
	GetServiceVersion(ctx context.Context, id int64) (ServiceVersion, error)
	GetServiceVersions(ctx context.Context, serviceID int64) ([]ServiceVersion, error)
	GetUser(ctx context.Context, id int64) (User, error)
	GetUserApplications(ctx context.Context, userID int64) ([]Application, error)
	GetUserServices(ctx context.Context, userID int64) ([]Service, error)
}

var _ Querier = (*Queries)(nil)
