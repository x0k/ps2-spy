package sql_unit_of_work

import (
	"context"
	"database/sql"
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/unit_of_work"
)

type UnitOfWork struct {
	tx *sql.Tx
}

func New(ctx context.Context, db *sql.DB) (UnitOfWork, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return UnitOfWork{}, err
	}
	return UnitOfWork{tx: tx}, nil
}

func (uow UnitOfWork) Tx() *sql.Tx {
	return uow.tx
}

func (uow UnitOfWork) Commit(ctx context.Context) error {
	return uow.tx.Commit()
}

func (uow UnitOfWork) Rollback(ctx context.Context) error {
	err := uow.tx.Rollback()
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}

func NewFactory(db *sql.DB) unit_of_work.Factory[*sql.Tx] {
	return func(ctx context.Context) (unit_of_work.UnitOfWork[*sql.Tx], error) {
		return New(ctx, db)
	}
}
