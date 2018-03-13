health
======

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
	    "error": "goroutine 23 [running]:\nsomepackage/somepackage/oops.func·001()..."
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
