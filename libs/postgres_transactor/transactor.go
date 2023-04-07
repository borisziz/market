package transactor

//go:generate sh -c "mkdir -p mocks && rm -rf mocks/manager_minimock.go"
//go:generate minimock -i github.com/jackc/pgx/v4.Tx -o ./mocks/tx_minimock.go

import (
	"context"
	"route256/libs/sqlmetrics"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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
	err = prometheus.Register(sqlmetrics.NewCollector(pool, config.ConnConfig.Database))
	if err != nil {
		return nil, errors.Wrap(err, "register collector")
	}
	return &TransactionManager{
		pool: pool,
	}, nil
}

type TxKey string

const key = TxKey("tx")

func (tm *TransactionManager) RunTransaction(ctx context.Context, isoLevel string, fx func(ctxTX context.Context) error) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "transaction")
	defer span.Finish()
	span.SetTag("level", isoLevel)

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
		ext.Error.Set(span, true)
		return err
	}
	if err := fx(context.WithValue(ctx, key, tx)); err != nil {
		ext.Error.Set(span, true)
		return multierr.Combine(err, tx.Rollback(ctx))
	}
	if err := tx.Commit(ctx); err != nil {
		ext.Error.Set(span, true)
		return multierr.Combine(err, tx.Rollback(ctx))
	}

	return nil
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(key).(QueryEngine)
	if ok && tx != nil {
		return sqlmetrics.NewQueryEngine(tx, tm.pool.Config().ConnConfig.Database)
	}
	return sqlmetrics.NewQueryEngine(tm.pool, tm.pool.Config().ConnConfig.Database)
}
