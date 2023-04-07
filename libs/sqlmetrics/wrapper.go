package sqlmetrics

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	operationStatusSuccess = "true"
	operationStatusFailed  = "false"
	spanNameQuery          = "query"
	spanNameExec           = "exec"
	tagQuery               = "query"
	tagArgs                = "args"
)

type QueryEngine interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
}

type QueryEngineWrapper struct {
	realEngine QueryEngine
	dbName     string
}

func NewQueryEngine(db QueryEngine, dbName string) *QueryEngineWrapper {
	return &QueryEngineWrapper{realEngine: db, dbName: dbName}
}

func (db *QueryEngineWrapper) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, spanNameQuery)
	defer span.Finish()
	span.SetTag(tagQuery, query)
	span.SetTag(tagArgs, args)

	timeStart := time.Now()
	rows, err := db.realEngine.Query(ctx, query, args...)
	HistogramQueryTime.WithLabelValues(db.dbName, query).Observe(time.Since(timeStart).Seconds())
	if err != nil {
		QueryCounter.WithLabelValues(db.dbName, query, operationStatusFailed).Inc()
		ext.Error.Set(span, true)
	} else {
		QueryCounter.WithLabelValues(db.dbName, query, operationStatusSuccess).Inc()
	}
	return rows, err
}

func (db *QueryEngineWrapper) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, spanNameExec)
	defer span.Finish()

	span.SetTag(tagQuery, query)
	span.SetTag(tagArgs, args)

	timeStart := time.Now()
	cmd, err := db.realEngine.Exec(ctx, query, args...)
	HistogramExecTime.WithLabelValues(db.dbName, query).Observe(time.Since(timeStart).Seconds())
	if err != nil {
		QueryCounter.WithLabelValues(db.dbName, query, operationStatusFailed).Inc()
		ext.Error.Set(span, true)
	} else {
		QueryCounter.WithLabelValues(db.dbName, query, operationStatusSuccess).Inc()
	}
	return cmd, err
}

func (db *QueryEngineWrapper) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return db.realEngine.SendBatch(ctx, b)
}

func (db *QueryEngineWrapper) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.realEngine.Begin(ctx)
}
