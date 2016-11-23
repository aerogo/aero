package aero

import (
	"encoding/json"
	"runtime"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	humanize "github.com/dustin/go-humanize"
	"github.com/valyala/fasthttp"
)

// Route statistics
type Route struct {
	Route        string
	Requests     uint64
	ResponseTime uint64
}

// byResponseTime ...
type byResponseTime []*Route

func (c byResponseTime) Len() int {
	return len(c)
}

func (c byResponseTime) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c byResponseTime) Less(i, j int) bool {
	return c[i].ResponseTime > c[j].ResponseTime
}

// byRequests ...
type byRequests []*Route

func (c byRequests) Len() int {
	return len(c)
}

func (c byRequests) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c byRequests) Less(i, j int) bool {
	return c[i].Requests > c[j].Requests
}

// showStatistics ...
func (app *Application) showStatistics(path string) {
	// Statistics route
	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		avg := sigar.LoadAverage{}
		uptime := sigar.Uptime{}

		avg.Get()
		uptime.Get()

		mem := sigar.Mem{}
		mem.Get()

		type AppMemoryStats struct {
			Allocated   string
			GCThreshold string
			Objects     uint64
		}

		type SystemMemoryStats struct {
			Total string
			Free  string
			Cache string
		}

		type AppStats struct {
			Go       string
			Uptime   string
			Requests uint64
			Memory   AppMemoryStats
			Config   *Configuration
		}

		type SystemStats struct {
			Uptime      string
			CPUs        int
			LoadAverage sigar.LoadAverage
			Memory      SystemMemoryStats
		}

		type RouteSummary struct {
			Slow    []*Route
			Popular []*Route
		}

		routeSummary := RouteSummary{}

		for path, stats := range app.routeStatistics {
			route := &Route{
				Route:        path,
				Requests:     atomic.LoadUint64(&stats.requestCount),
				ResponseTime: uint64(stats.AverageResponseTime()),
			}

			if route.ResponseTime >= 10 {
				routeSummary.Slow = append(routeSummary.Slow, route)
			}

			if route.Requests >= 1 {
				routeSummary.Popular = append(routeSummary.Popular, route)
			}
		}

		sort.Sort(byResponseTime(routeSummary.Slow))
		sort.Sort(byRequests(routeSummary.Popular))

		stats := struct {
			System SystemStats
			App    AppStats
			Routes RouteSummary
		}{
			System: SystemStats{
				Uptime:      strings.TrimSpace(uptime.Format()),
				CPUs:        runtime.NumCPU(),
				LoadAverage: avg,
				Memory: SystemMemoryStats{
					Total: humanize.Bytes(mem.Total),
					Free:  humanize.Bytes(mem.Free),
					Cache: humanize.Bytes(mem.Used - mem.ActualUsed),
				},
			},
			App: AppStats{
				Go:       strings.Replace(runtime.Version(), "go", "", 1),
				Uptime:   strings.TrimSpace(humanize.RelTime(app.start, time.Now(), "", "")),
				Requests: app.RequestCount(),
				Memory: AppMemoryStats{
					Allocated:   humanize.Bytes(memStats.HeapAlloc),
					GCThreshold: humanize.Bytes(memStats.NextGC),
					Objects:     memStats.HeapObjects,
				},
				Config: app.Config,
			},
			Routes: routeSummary,
		}

		// numCPU :=
		// var b bytes.Buffer
		// b.WriteString("Server statistics:\n")

		// b.WriteString("\nGo version: ")
		// b.WriteString(runtime.Version())

		// b.WriteString("\nCPUs: ")
		// b.WriteString(strconv.Itoa(numCPU))

		fasthttpContext.Response.Header.Set("Content-Type", "application/json")
		bytes, err := json.Marshal(stats)
		if err != nil {
			fasthttpContext.WriteString("Error serializing to JSON")
			return
		}
		fasthttpContext.Write(bytes)
	})
}
