package storage

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/db"
)

type Storage interface {
	Queries() *db.Queries
	Transaction(context.Context, func(s Storage) error) error
}
