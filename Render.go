package aero

import (
	"encoding/json"
	"runtime"
	"strconv"

	"github.com/OneOfOne/xxhash"
	cache "github.com/patrickmn/go-cache"
	"github.com/robertkrimen/otto"
)

var renderJobs chan renderJob
var renderResults chan string

type renderJob struct {
	template *Template
	params   map[string]interface{}
}

func init() {
	renderJobs = make(chan renderJob, 4096)
	renderResults = make(chan string, 4096)

	for w := 1; w <= runtime.NumCPU(); w++ {
		go worker(renderJobs, renderResults)
	}
}

func worker(jobs <-chan renderJob, results chan<- string) {
	vm := otto.New()

	for job := range jobs {
		h := xxhash.NewS64(0)

		for _, value := range job.params {
			jsonBytes, err := json.Marshal(value)

			if err == nil {
				h.Write(jsonBytes)
			}
		}

		hash := strconv.FormatUint(h.Sum64(), 10)
		cachedResponse, found := job.template.renderCache.Get(hash)

		if found {
			results <- cachedResponse.(string)
		} else {
			for key, value := range job.params {
				vm.Set(key, value)
			}

			result, _ := vm.Run(job.template.Script)
			code, _ := result.ToString()
			results <- code

			job.template.renderCache.Set(hash, code, cache.DefaultExpiration)
		}
	}
}
