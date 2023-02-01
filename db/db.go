// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.createServiceStmt, err = db.PrepareContext(ctx, createService); err != nil {
		return nil, fmt.Errorf("error preparing query CreateService: %w", err)
	}
	if q.createServiceVersionStmt, err = db.PrepareContext(ctx, createServiceVersion); err != nil {
		return nil, fmt.Errorf("error preparing query CreateServiceVersion: %w", err)
	}
	if q.createUserStmt, err = db.PrepareContext(ctx, createUser); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUser: %w", err)
	}
	if q.getServiceVersionStmt, err = db.PrepareContext(ctx, getServiceVersion); err != nil {
		return nil, fmt.Errorf("error preparing query GetServiceVersion: %w", err)
	}
	if q.getServiceVersionsStmt, err = db.PrepareContext(ctx, getServiceVersions); err != nil {
		return nil, fmt.Errorf("error preparing query GetServiceVersions: %w", err)
	}
	if q.getUserStmt, err = db.PrepareContext(ctx, getUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetUser: %w", err)
	}
	if q.getUserApplicationsStmt, err = db.PrepareContext(ctx, getUserApplications); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserApplications: %w", err)
	}
	if q.getUserServicesStmt, err = db.PrepareContext(ctx, getUserServices); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserServices: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.createServiceStmt != nil {
		if cerr := q.createServiceStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createServiceStmt: %w", cerr)
		}
	}
	if q.createServiceVersionStmt != nil {
		if cerr := q.createServiceVersionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createServiceVersionStmt: %w", cerr)
		}
	}
	if q.createUserStmt != nil {
		if cerr := q.createUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserStmt: %w", cerr)
		}
	}
	if q.getServiceVersionStmt != nil {
		if cerr := q.getServiceVersionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getServiceVersionStmt: %w", cerr)
		}
	}
	if q.getServiceVersionsStmt != nil {
		if cerr := q.getServiceVersionsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getServiceVersionsStmt: %w", cerr)
		}
	}
	if q.getUserStmt != nil {
		if cerr := q.getUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserStmt: %w", cerr)
		}
	}
	if q.getUserApplicationsStmt != nil {
		if cerr := q.getUserApplicationsStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserApplicationsStmt: %w", cerr)
		}
	}
	if q.getUserServicesStmt != nil {
		if cerr := q.getUserServicesStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserServicesStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db                       DBTX
	tx                       *sql.Tx
	createServiceStmt        *sql.Stmt
	createServiceVersionStmt *sql.Stmt
	createUserStmt           *sql.Stmt
	getServiceVersionStmt    *sql.Stmt
	getServiceVersionsStmt   *sql.Stmt
	getUserStmt              *sql.Stmt
	getUserApplicationsStmt  *sql.Stmt
	getUserServicesStmt      *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                       tx,
		tx:                       tx,
		createServiceStmt:        q.createServiceStmt,
		createServiceVersionStmt: q.createServiceVersionStmt,
		createUserStmt:           q.createUserStmt,
		getServiceVersionStmt:    q.getServiceVersionStmt,
		getServiceVersionsStmt:   q.getServiceVersionsStmt,
		getUserStmt:              q.getUserStmt,
		getUserApplicationsStmt:  q.getUserApplicationsStmt,
		getUserServicesStmt:      q.getUserServicesStmt,
	}
}
