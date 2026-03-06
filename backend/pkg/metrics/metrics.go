// Package metrics 提供 Prometheus 监控指标
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal HTTP 请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration HTTP 请求延迟
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestSize HTTP 请求大小
	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// HTTPResponseSize HTTP 响应大小
	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// DBQueryDuration 数据库查询延迟
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// DBConnectionsActive 活跃数据库连接数
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// CacheHitRatio 缓存命中率
	CacheHitRatio = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"result"},
	)

	// BusinessOperationsTotal 业务操作计数
	BusinessOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "resource", "status"},
	)

	// ActiveUsers 活跃用户数
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users",
		},
	)
)

// RecordHTTPRequest 记录 HTTP 请求指标
func RecordHTTPRequest(method, path, status string, duration float64, reqSize, respSize int64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	if reqSize > 0 {
		HTTPRequestSize.WithLabelValues(method, path).Observe(float64(reqSize))
	}
	if respSize > 0 {
		HTTPResponseSize.WithLabelValues(method, path).Observe(float64(respSize))
	}
}

// RecordDBQuery 记录数据库查询指标
func RecordDBQuery(operation, table string, duration float64) {
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit() {
	CacheHitRatio.WithLabelValues("hit").Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss() {
	CacheHitRatio.WithLabelValues("miss").Inc()
}

// RecordBusinessOperation 记录业务操作
func RecordBusinessOperation(operation, resource, status string) {
	BusinessOperationsTotal.WithLabelValues(operation, resource, status).Inc()
}

// SetDBConnections 设置数据库连接数
func SetDBConnections(count float64) {
	DBConnectionsActive.Set(count)
}

// SetActiveUsers 设置活跃用户数
func SetActiveUsers(count float64) {
	ActiveUsers.Set(count)
}
