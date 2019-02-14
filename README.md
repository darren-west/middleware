# Middleware [![Build Status](https://travis-ci.org/darren-west/middleware.svg?branch=master)](https://travis-ci.org/darren-west/middleware)

This is a simple middleware library for GoLangs http.Handler. It is designed to be bloat free and have a simple API for building a stack of middleware. 

### Getting started
A simple example is below

```go
	m, err := middleware.New(
		middleware.WithFunc(func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			next(w, r)
		}),
		middleware.UseHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "middleware!")
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":80", m))
```

### TODO
* Document code
* Finalize on API - possible use Builder struct to make New call cleaner.
* Improve tests (possibly use suites and or table driven tests)
* Benchmark tests
* Look for improvements in performance and remove redundant code
* More example - Runnable examples.