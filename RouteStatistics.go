package aero

import "sync/atomic"

// RouteStatistics includes performance statistics for a specific route.
type RouteStatistics struct {
	requestCount uint64
	responseTime uint64
}

// AverageResponseTime returns the average response time of the route.
func (stats *RouteStatistics) AverageResponseTime() float64 {
	requestCount := atomic.LoadUint64(&stats.requestCount)
	responseTime := atomic.LoadUint64(&stats.responseTime)

	if requestCount == 0 {
		return 0
	}

	return float64(responseTime) / float64(requestCount)
}
