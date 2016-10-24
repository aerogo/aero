package aero

import (
	"runtime"
	"strconv"

	"github.com/OneOfOne/xxhash"
	cache "github.com/patrickmn/go-cache"
	"github.com/robertkrimen/otto"
	"gopkg.in/mgo.v2/bson"
)

// Map is equivalent to map[string]interface{}.
type Map map[string]interface{}

var renderJobs chan renderJob
var renderResults chan string

type renderJob struct {
	template *Template
	params   Map
}

func init() {
	renderJobs = make(chan renderJob, 4096)
	renderResults = make(chan string, 4096)

	for w := 1; w <= runtime.NumCPU(); w++ {
		go renderWorker(renderJobs, renderResults)
	}
}

func renderWorker(jobs <-chan renderJob, results chan<- string) {
	vm := otto.New()

	for job := range jobs {
		h := xxhash.NewS64(0)
		serialized, _ := bson.Marshal(job.params)
		h.Write(serialized)

		hash := strconv.FormatUint(h.Sum64(), 10)
		cachedResponse, found := job.template.renderCache.Get(hash)

		if found {
			results <- cachedResponse.(string)
		} else {
			for key, value := range job.params {
				vm.Set(key, value)
			}

			if job.template.Script == nil {
				results <- job.template.syntaxError
				continue
			}

			result, runtimeError := vm.Run(job.template.Script)

			if runtimeError != nil {
				results <- runtimeError.Error()
				continue
			}

			code, toStringError := result.ToString()

			if toStringError != nil {
				results <- toStringError.Error()
				continue
			}

			results <- code

			job.template.renderCache.Set(hash, code, cache.DefaultExpiration)
		}
	}
}
