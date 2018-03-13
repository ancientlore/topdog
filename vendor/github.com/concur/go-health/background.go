package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

const (
	// DefaultFrequency is how often background tests will run when not specified.
	DefaultFrequency = 1 * time.Minute
)

// Ticker defines information for background tests.
type Ticker struct {
	Tester
	Frequency time.Duration // How often to run the tests
	results   Results       // results of tests
	lock      sync.RWMutex
	cancel    context.CancelFunc
}

// runOnce runs the tests once
func (tick *Ticker) runOnce() {
	r := tick.Run()
	tick.lock.Lock()
	tick.results = r
	tick.lock.Unlock()
}

// run runs the background health check at the configured interval
func (tick *Ticker) run(ctx context.Context) {
	freq := tick.Frequency
	if freq <= 0 {
		freq = DefaultFrequency
	}
	tck := time.NewTicker(freq)
	done := ctx.Done()
	for {
		select {
		case <-tck.C:
			tick.runOnce()
		case <-done:
			tck.Stop()
			return
		}
	}
}

// Start starts the background health check.
func (tick *Ticker) Start() {
	tick.Stop()
	ctx := tick.Context
	if ctx == nil {
		ctx = context.Background()
	}
	tick.lock.Lock()
	ctx, tick.cancel = context.WithCancel(ctx)
	go tick.run(ctx)
	tick.lock.Unlock()
}

// Stop stops the background health check.
func (tick *Ticker) Stop() {
	tick.lock.Lock()
	if tick.cancel != nil {
		tick.cancel()
		tick.cancel = nil
	}
	tick.lock.Unlock()
}

// GetResults returns the current results
func (tick *Ticker) GetResults() Results {
	tick.lock.RLock()
	defer tick.lock.RUnlock()
	// need to copy for thread safety
	r := make(Results)
	for k, v := range tick.results {
		r[k] = v
	}
	return r
}

// ServeHTTP serves requests by running all the tests and returning a JSON block with the results.
// If all the tests succeed, a 200 HTTP status is returned. Otherwise, a 500 HTTP status is returned.
func (tick *Ticker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tick.lock.RLock()
	defer tick.lock.RUnlock()
	res := tick.results
	if res == nil {
		res = make(Results)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if res.Failed() {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	b, err := json.Marshal(res)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(b)
	}
}
