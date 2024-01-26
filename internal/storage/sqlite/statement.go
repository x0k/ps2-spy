package sqlite

import (
	"context"
	"database/sql"
)

type row interface {
	Err() error
	Scan(dest ...any) error
}

type statement interface {
	UseTx(ctx context.Context, tx *sql.Tx) statement
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, args ...any) row
	QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
	Close() error
}

type simpleStatement struct {
	*sql.Stmt
}

func (s simpleStatement) UseTx(ctx context.Context, tx *sql.Tx) statement {
	return simpleStatement{tx.StmtContext(ctx, s.Stmt)}
}

func (s simpleStatement) QueryRowContext(ctx context.Context, args ...any) row {
	return s.Stmt.QueryRowContext(ctx, args...)
}

type errRow struct {
	err error
}

func (r errRow) Err() error {
	return r.err
}

func (r errRow) Scan(dest ...any) error {
	return r.err
}

type dynamicStatement struct {
	db           *sql.DB
	queryBuilder func(args ...any) (string, error)
	tx           *sql.Tx
}

func (s *dynamicStatement) UseTx(_ context.Context, tx *sql.Tx) statement {
	s.tx = tx
	return s
}

func (s *dynamicStatement) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	query, err := s.queryBuilder(args...)
	if err != nil {
		return nil, err
	}
	if s.tx != nil {
		return s.tx.ExecContext(ctx, query, args...)
	}
	return s.db.ExecContext(ctx, query, args...)
}

func (s *dynamicStatement) QueryRowContext(ctx context.Context, args ...any) row {
	query, err := s.queryBuilder(args...)
	if err != nil {
		return errRow{err}
	}
	if s.tx != nil {
		return s.tx.QueryRowContext(ctx, query, args...)
	}
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *dynamicStatement) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	query, err := s.queryBuilder(args...)
	if err != nil {
		return nil, err
	}
	if s.tx != nil {
		return s.tx.QueryContext(ctx, query, args...)
	}
	return s.db.QueryContext(ctx, query, args...)
}

func (s *dynamicStatement) Close() error {
	s.tx = nil
	return nil
}
