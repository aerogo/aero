package aero

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	humanize "github.com/dustin/go-humanize"
	"github.com/valyala/fasthttp"
)

func (app *Application) showStatistics(path string) {
	// Statistics route
	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		monitor := sigar.ConcreteSigar{}
		uptime := sigar.Uptime{}

		avg, _ := monitor.GetLoadAverage()
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
			Config   Configuration
		}

		type SystemStats struct {
			Uptime      string
			CPUs        int
			LoadAverage sigar.LoadAverage
			Memory      SystemMemoryStats
		}

		stats := struct {
			System SystemStats
			App    AppStats
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
				Requests: atomic.LoadUint64(&app.requestCount),
				Memory: AppMemoryStats{
					Allocated:   humanize.Bytes(memStats.HeapAlloc),
					GCThreshold: humanize.Bytes(memStats.NextGC),
					Objects:     memStats.HeapObjects,
				},
				Config: app.Config,
			},
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
