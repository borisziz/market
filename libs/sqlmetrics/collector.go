package sqlmetrics

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PoolStatsCollector struct {
	db *pgxpool.Pool

	acquireCount            *prometheus.Desc
	acquireDuration         *prometheus.Desc
	acquireConns            *prometheus.Desc
	canceledAcquireCount    *prometheus.Desc
	constructingConns       *prometheus.Desc
	emptyAcquireCount       *prometheus.Desc
	idleConns               *prometheus.Desc
	maxConns                *prometheus.Desc
	totalConns              *prometheus.Desc
	newConnsCount           *prometheus.Desc
	maxLifetimeDestroyCount *prometheus.Desc
	maxIdleDestroyCount     *prometheus.Desc
}

func NewCollector(db *pgxpool.Pool, dbName string) *PoolStatsCollector {
	return &PoolStatsCollector{
		db: db,
		acquireCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "acquire_count"),
			"The cumulative count of successful acquires from the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		acquireDuration: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "acquire_duration"),
			"The total duration of all successful acquires from the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		acquireConns: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "acquire_conns"),
			"The number of currently acquired connections in the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		canceledAcquireCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "canceled_acquire_count"),
			"The cumulative count of acquires from the pool that were canceled by a context",
			nil,
			prometheus.Labels{"db": dbName},
		),
		constructingConns: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "constructing_conns"),
			"The number of conns with construction in progress in the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		emptyAcquireCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "empty_acquire_count"),
			" The cumulative count of successful acquires from the pool that waited "+
				"for a resource to be released or constructed because the pool was empty.",
			nil,
			prometheus.Labels{"db": dbName},
		),
		idleConns: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "idle_conns"),
			"The number of currently idle conns in the pool.",
			nil,
			prometheus.Labels{"db": dbName},
		),
		maxConns: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "max_conns"),
			"The maximum size of the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		totalConns: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "total_conns"),
			"The total number of resources currently in the pool",
			nil,
			prometheus.Labels{"db": dbName},
		),
		newConnsCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "new_conns_count"),
			"The cumulative count of new connections opened",
			nil,
			prometheus.Labels{"db": dbName},
		),
		maxLifetimeDestroyCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "max_lifetime_destroy_count"),
			"The cumulative count of connections destroyed because they exceeded MaxConnLifetime",
			nil,
			prometheus.Labels{"db": dbName},
		),
		maxIdleDestroyCount: prometheus.NewDesc(
			prometheus.BuildFQName("homework", "pool", "max_idle_destroy_count"),
			"The cumulative count of connections destroyed because they exceeded MaxConnIdleTime",
			nil,
			prometheus.Labels{"db": dbName},
		),
	}
}

func (p PoolStatsCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- p.acquireCount
	descs <- p.acquireDuration
	descs <- p.acquireConns
	descs <- p.canceledAcquireCount
	descs <- p.constructingConns
	descs <- p.emptyAcquireCount
	descs <- p.idleConns
	descs <- p.maxConns
	descs <- p.totalConns
	descs <- p.newConnsCount
	descs <- p.maxLifetimeDestroyCount
	descs <- p.maxIdleDestroyCount
}

func (p PoolStatsCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := p.db.Stat()
	metrics <- prometheus.MustNewConstMetric(p.acquireCount, prometheus.CounterValue, float64(stats.AcquireCount()))
	metrics <- prometheus.MustNewConstMetric(p.acquireDuration, prometheus.CounterValue, float64(stats.AcquireDuration()))
	metrics <- prometheus.MustNewConstMetric(p.acquireConns, prometheus.GaugeValue, float64(stats.AcquiredConns()))
	metrics <- prometheus.MustNewConstMetric(p.canceledAcquireCount, prometheus.CounterValue, float64(stats.CanceledAcquireCount()))
	metrics <- prometheus.MustNewConstMetric(p.constructingConns, prometheus.GaugeValue, float64(stats.ConstructingConns()))
	metrics <- prometheus.MustNewConstMetric(p.emptyAcquireCount, prometheus.CounterValue, float64(stats.EmptyAcquireCount()))
	metrics <- prometheus.MustNewConstMetric(p.idleConns, prometheus.GaugeValue, float64(stats.IdleConns()))
	metrics <- prometheus.MustNewConstMetric(p.maxConns, prometheus.GaugeValue, float64(stats.MaxConns()))
	metrics <- prometheus.MustNewConstMetric(p.totalConns, prometheus.GaugeValue, float64(stats.TotalConns()))
	metrics <- prometheus.MustNewConstMetric(p.newConnsCount, prometheus.CounterValue, float64(stats.NewConnsCount()))
	metrics <- prometheus.MustNewConstMetric(p.maxLifetimeDestroyCount, prometheus.CounterValue, float64(stats.MaxLifetimeDestroyCount()))
	metrics <- prometheus.MustNewConstMetric(p.maxIdleDestroyCount, prometheus.CounterValue, float64(stats.MaxIdleDestroyCount()))
}

var (
	QueryCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "homework",
		Subsystem: "pool",
		Name:      "query_total",
	},
		[]string{"db", "query", "success"},
	)
	ExecCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "homework",
		Subsystem: "pool",
		Name:      "exec_total",
	},
		[]string{"db", "query", "success"},
	)
	HistogramQueryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "homework",
		Subsystem: "pool",
		Name:      "histogram_query_time_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
	},
		[]string{"db", "query"},
	)
	HistogramExecTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "loms",
		Subsystem: "pool",
		Name:      "histogram_exec_time_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
	},
		[]string{"db", "query"},
	)
)
