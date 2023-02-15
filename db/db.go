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
	if q.createApplicationStmt, err = db.PrepareContext(ctx, createApplication); err != nil {
		return nil, fmt.Errorf("error preparing query CreateApplication: %w", err)
	}
	if q.createLoginTokenStmt, err = db.PrepareContext(ctx, createLoginToken); err != nil {
		return nil, fmt.Errorf("error preparing query CreateLoginToken: %w", err)
	}
	if q.createUserStmt, err = db.PrepareContext(ctx, createUser); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUser: %w", err)
	}
	if q.createUserSessionStmt, err = db.PrepareContext(ctx, createUserSession); err != nil {
		return nil, fmt.Errorf("error preparing query CreateUserSession: %w", err)
	}
	if q.deleteLoginTokenStmt, err = db.PrepareContext(ctx, deleteLoginToken); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteLoginToken: %w", err)
	}
	if q.deleteUserSessionStmt, err = db.PrepareContext(ctx, deleteUserSession); err != nil {
		return nil, fmt.Errorf("error preparing query DeleteUserSession: %w", err)
	}
	if q.getApplicationStmt, err = db.PrepareContext(ctx, getApplication); err != nil {
		return nil, fmt.Errorf("error preparing query GetApplication: %w", err)
	}
	if q.getLoginTokenStmt, err = db.PrepareContext(ctx, getLoginToken); err != nil {
		return nil, fmt.Errorf("error preparing query GetLoginToken: %w", err)
	}
	if q.getUserStmt, err = db.PrepareContext(ctx, getUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetUser: %w", err)
	}
	if q.getUserApplicationsStmt, err = db.PrepareContext(ctx, getUserApplications); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserApplications: %w", err)
	}
	if q.getUserByEmailOrUsernameStmt, err = db.PrepareContext(ctx, getUserByEmailOrUsername); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserByEmailOrUsername: %w", err)
	}
	if q.getUserSessionStmt, err = db.PrepareContext(ctx, getUserSession); err != nil {
		return nil, fmt.Errorf("error preparing query GetUserSession: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.createApplicationStmt != nil {
		if cerr := q.createApplicationStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createApplicationStmt: %w", cerr)
		}
	}
	if q.createLoginTokenStmt != nil {
		if cerr := q.createLoginTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createLoginTokenStmt: %w", cerr)
		}
	}
	if q.createUserStmt != nil {
		if cerr := q.createUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserStmt: %w", cerr)
		}
	}
	if q.createUserSessionStmt != nil {
		if cerr := q.createUserSessionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing createUserSessionStmt: %w", cerr)
		}
	}
	if q.deleteLoginTokenStmt != nil {
		if cerr := q.deleteLoginTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteLoginTokenStmt: %w", cerr)
		}
	}
	if q.deleteUserSessionStmt != nil {
		if cerr := q.deleteUserSessionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing deleteUserSessionStmt: %w", cerr)
		}
	}
	if q.getApplicationStmt != nil {
		if cerr := q.getApplicationStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getApplicationStmt: %w", cerr)
		}
	}
	if q.getLoginTokenStmt != nil {
		if cerr := q.getLoginTokenStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getLoginTokenStmt: %w", cerr)
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
	if q.getUserByEmailOrUsernameStmt != nil {
		if cerr := q.getUserByEmailOrUsernameStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserByEmailOrUsernameStmt: %w", cerr)
		}
	}
	if q.getUserSessionStmt != nil {
		if cerr := q.getUserSessionStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserSessionStmt: %w", cerr)
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
	db                           DBTX
	tx                           *sql.Tx
	createApplicationStmt        *sql.Stmt
	createLoginTokenStmt         *sql.Stmt
	createUserStmt               *sql.Stmt
	createUserSessionStmt        *sql.Stmt
	deleteLoginTokenStmt         *sql.Stmt
	deleteUserSessionStmt        *sql.Stmt
	getApplicationStmt           *sql.Stmt
	getLoginTokenStmt            *sql.Stmt
	getUserStmt                  *sql.Stmt
	getUserApplicationsStmt      *sql.Stmt
	getUserByEmailOrUsernameStmt *sql.Stmt
	getUserSessionStmt           *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:                           tx,
		tx:                           tx,
		createApplicationStmt:        q.createApplicationStmt,
		createLoginTokenStmt:         q.createLoginTokenStmt,
		createUserStmt:               q.createUserStmt,
		createUserSessionStmt:        q.createUserSessionStmt,
		deleteLoginTokenStmt:         q.deleteLoginTokenStmt,
		deleteUserSessionStmt:        q.deleteUserSessionStmt,
		getApplicationStmt:           q.getApplicationStmt,
		getLoginTokenStmt:            q.getLoginTokenStmt,
		getUserStmt:                  q.getUserStmt,
		getUserApplicationsStmt:      q.getUserApplicationsStmt,
		getUserByEmailOrUsernameStmt: q.getUserByEmailOrUsernameStmt,
		getUserSessionStmt:           q.getUserSessionStmt,
	}
}
