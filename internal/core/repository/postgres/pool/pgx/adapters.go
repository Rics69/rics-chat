package core_pgx_pool

import (
	"errors"
	"fmt"

	core_postgres_pool "github.com/Rics69/rics-chat/internal/core/repository/postgres/pool"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgxRows struct {
	pgx.Rows
}

func (r pgxRows) Scan(dest ...any) error {
	return MapErrors(r.Rows.Scan(dest...))
}

func (r pgxRows) Err() error {
	return MapErrors(r.Rows.Err())
}

type pgxRow struct {
	pgx.Row
}

func (r pgxRow) Scan(dest ...any) error {
	return MapErrors(r.Row.Scan(dest...))
}

type pgxCommandTag struct {
	pgconn.CommandTag
}

// MapErrors переводит ошибки pgx в ошибки нашего пула,
// чтобы репозитории не зависели от конкретного драйвера.
func MapErrors(err error) error {
	const (
		pgxViolatesForeignKeyErrorCode = "23503"
	)

	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return core_postgres_pool.ErrNoRows
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgxViolatesForeignKeyErrorCode:
			return fmt.Errorf("%w: %w", err, core_postgres_pool.ErrViolatesForeignKey)
		}
	}

	// два %w (Go 1.20+): сохраняем и исходную ошибку (например context.DeadlineExceeded),
	// и нашу метку ErrUnknown — errors.Is сработает для обеих
	return fmt.Errorf("%w: %w", err, core_postgres_pool.ErrUnknown)
}
