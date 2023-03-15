package transactor

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type QueryEngine interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func New(connectString string) (*TransactionManager, error) {
	config, err := pgxpool.ParseConfig(connectString)
	if err != nil {
		return nil, errors.Wrap(err, "parse connect")
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, errors.Wrap(err, "connect")
	}
	return &TransactionManager{
		pool: pool,
	}, nil
}

type txKey string

const key = txKey("tx")

func (tm *TransactionManager) RunTransaction(ctx context.Context, isoLevel string, fx func(ctxTX context.Context) error) error {
	var txIsoLevel pgx.TxIsoLevel
	switch isoLevel {
	case "serializable":
		txIsoLevel = pgx.Serializable
	case "repeatable read":
		txIsoLevel = pgx.RepeatableRead
	case "read committed":
		txIsoLevel = pgx.ReadCommitted
	case "read uncommitted":
		txIsoLevel = pgx.ReadUncommitted
	default:
		txIsoLevel = pgx.Serializable
	}
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: txIsoLevel,
	})
	if err != nil {
		return err
	}
	if err := fx(context.WithValue(ctx, key, tx)); err != nil {
		return multierr.Combine(err, tx.Rollback(ctx))
	}
	if err := tx.Commit(ctx); err != nil {
		return multierr.Combine(err, tx.Rollback(ctx))
	}

	return nil
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(key).(QueryEngine)
	if ok && tx != nil {
		return tx
	}

	return tm.pool
}
