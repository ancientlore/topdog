/*
Package health is a framework for running health tests. Test functions are
run in parallel and results can be reported via a web handler.

To get started, initialize the Tests map with named tests:

	tester = health.Tester{
		Tests: health.TestFuncs{
			"myTest": func(ctx context.Context) error {
				// ctx.Done() is a channel that will report if the timeout was exceeded
				// or if processing should be cancelled
				// do some test
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

You can manually invoke tests using Run:

	var results health.Results = tester.Run()

Or, use the HTTP handler and have a load balancer periodically run and check the results:

	http.Handle("/health", tester)

If all the tests succeed, an HTTP 200 is returned. Otherwise, and HTTP 500 is returned. Both
cases return JSON:

	{
	  "database": {
	    "healthy": true
	  },
	  "memcached": {
	    "healthy": true
	  },
	  "logic": {
	    "healthy": false,
	    "message": "OH. MY. GOD.",
	    "error": "goroutine 23 [running]:\nsomepackage/somepackage.git/oops.func·001()..."
	  }
	}

Note that all tests are run in parallel, and the system includes code to trap calls to panic().
Tests should respect the timeout by checking ctx.Done(), however the system will not break if they
don't check.

Tests can also be run periodically in the background using the Ticker type. The difference is that a
thread is started to periodically run the tests, and a mechanism is provided to get the results.

Start by defining some tests:

	bgtester = .Ticker{
		Tester: health.Tester{
			Tests: health.TestFuncs{
				"myTest": func(ctx context.Context) error {
					// ctx.Done() is a channel that will report if the timeout was exceeded
					// or if processing should be cancelled
					// do some test
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

Then start the background testing (and provide a means for it to stop running at some point,
usually the end of the program):

	bg.Start()
	defer bg.Stop()

You can manually check the results using GetResults:

	var results health.Results = bg.GetResults()

Or, use the HTTP handler and have a load balancer periodically check the results:

	http.Handle("/health", &bg)

If all the tests succeed, an HTTP 200 is returned. Otherwise, and HTTP 500 is returned. Both
cases return JSON.
*/
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// A Result holds the results of a single test.
type Result struct {
	Healthy bool   `json:"healthy"`           // Whether this part of the service is healthy.
	Message string `json:"message,omitempty"` // A message indicating what went wrong.
	Error   string `json:"error,omitempty"`   // Error or stack trace information, if available.
}

// Results maps test names to their results.
type Results map[string]Result

// TestFunc defines the type of a test function. Test functions receive a Context which has
// a timeout. Test functions can (and should) check the context's Done() channel, and stop
// their test if the deadline is reached.
type TestFunc func(context.Context) error

// LoggerFunc defines a function used to log messages when tests fail.
type LoggerFunc func(testName, messageText, errorText string)

// TestFuncs maps test names to their test functions.
type TestFuncs map[string]TestFunc

// Tester is used to invoke test functions, gather results, and provide HTTP access. Only the Tests
// member must be initialized.
type Tester struct {
	Timeout time.Duration   // The time that all the tests can take
	Context context.Context // The default context passed to the test functions; defaults to context.Background()
	Tests   TestFuncs       // The slice for storing the test methods to invoke
	Log     LoggerFunc      // If not nil, will be used to log messages when tests fail
}

const (
	// DefaultTimeout is how long tests can take when a timeout is not specified.
	DefaultTimeout = 2 * time.Second
)

// tp is used internally to communicate data over a channel.
type tp struct {
	name   string
	result *Result
}

// Failed returns true if any of the tests have failed.
func (r Results) Failed() bool {
	for _, x := range r {
		if x.Healthy != true {
			return true
		}
	}
	return false
}

// ServeHTTP serves requests by running all the tests and returning a JSON block with the results.
// If all the tests succeed, a 200 HTTP status is returned. Otherwise, a 500 HTTP status is returned.
func (t Tester) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	results := t.Run()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if results.Failed() {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	b, err := json.Marshal(results)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(b)
	}
}

// Run runs all of the tests in parallel and collects the results. Run provides a timeout and will
// return when the timeout is reached, even if some of the test functions are not complete. Test
// functions should check the context's Done() channel and stop if the test should be aborted.
// Run will handle panic() calls and errors from the test functions. You should not add tests
// while Run is active.
func (t Tester) Run() Results {
	var results = make(Results)
	if len(t.Tests) > 0 {
		rc := make(chan tp)
		timeout := t.Timeout
		if timeout <= 0 {
			timeout = DefaultTimeout
		}
		var cancel context.CancelFunc
		ctx := t.Context
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		for k, f := range t.Tests {
			go func(c context.Context, name string, fun TestFunc, ch chan<- tp) {
				defer func() {
					if err := recover(); err != nil {
						stack := make([]byte, 1024*8)
						stack = stack[:runtime.Stack(stack, false)]
						var desc string
						switch err.(type) {
						case error:
							desc = err.(error).Error()
						case string:
							desc = err.(string)
						default:
							desc = "PANIC"
						}
						ch <- tp{name: name, result: &Result{Healthy: false, Message: desc, Error: string(stack)}}
					}
				}()
				err := fun(c)
				if err != nil {
					ch <- tp{name: name, result: &Result{Healthy: false, Message: err.Error()}}
				} else {
					ch <- tp{name: name, result: &Result{Healthy: true}}
				}
			}(ctx, k, f, rc)
		}
		done := ctx.Done()
		for count := 0; count < len(t.Tests); {
			select {
			case r := <-rc:
				count++
				results[r.name] = *r.result
				if !r.result.Healthy && t.Log != nil {
					t.Log(r.name, r.result.Message, r.result.Error)
				}
			case <-done:
				count = len(t.Tests)
				for k2 := range t.Tests {
					_, ok := results[k2]
					if !ok {
						results[k2] = Result{Healthy: false, Message: ctx.Err().Error()}
						if t.Log != nil {
							t.Log(k2, ctx.Err().Error(), "")
						}
					}
				}
			}
		}
	}

	return results
}
