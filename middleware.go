// Package middleware provides a utility http.Handler for implementing middleware for net/http.
// The package includes no default middleware and is strictly for stitching middleware to a http handler.
package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
)

var (
	_ http.Handler = &Runner{}
)

//go:generate mockgen -destination ./mocks/http_handler.go -mock_names Handler=MockHTTPHandler -package mocks net/http Handler

//go:generate mockgen -destination ./mocks/middleware_handler.go -package mocks github.com/darren-west/middleware Handler

type (
	// Runner implements the http.Handler interface. It runs all the middleware set on instanation (using the New() func call)
	// before delegating to the destination http.Handler. If a piece of middleware does not call Next() the next middleware
	// will not be invoked.
	Runner struct {
		options *Options
	}

	// Next is a function for invoking the next middleware.
	Next http.HandlerFunc

	// Handler is an interface defining a piece of middleware. Middleware is similar to http.Handler but takes
	// a Next function for invoking (or not) the next piece of middleware.
	Handler interface {
		ServeHTTP(w http.ResponseWriter, r *http.Request, next Next)
	}

	// HandlerFunc is a convience function for implementing the Handler interface.
	HandlerFunc func(w http.ResponseWriter, r *http.Request, next Next)

	// HandlerIterator provides convience functions that operate on a slice of type Handler.
	HandlerIterator []Handler

	// Options holds the settable configuration for a Runner Handler.
	Options struct {
		Middleware HandlerIterator
		Handler    http.Handler
	}

	// Option allow the setting of options in the Runner Handler.
	Option func(*Options) error
)

// Options returns the Runner configuration options.
func (h *Runner) Options() Options {
	return *h.options
}

// ServeHTTP implements the http.Handler interface. It performs the stitching of the Middleware and invoking of
// the destination Handler.
func (h *Runner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	last := h.options.Handler
	for i := h.options.Middleware.Count() - 1; i >= 0; i-- {
		last = func(mid Handler, next http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				mid.ServeHTTP(w, r, next.ServeHTTP)
			}
		}(h.options.Middleware[i], last)
	}
	last.ServeHTTP(w, r)
}

// New allows instantiation a Runner Handler with some configuration.
// i.e. The UseHandlerFunc option allows the destination Handler to be set.
func New(setters ...Option) (*Runner, error) {
	opts, err := newOptions(setters...)
	if err != nil {
		return nil, err
	}
	return &Runner{
		options: opts,
	}, nil
}

// ServeHTTP implements the Handler interface for the HandlerFunc type.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next Next) {
	f(w, r, next)
}

func newOptions(optsetters ...Option) (*Options, error) {
	opts := &Options{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintf(w, "hello world!")
		}),
		Middleware: []Handler{},
	}
	for _, o := range optsetters {
		if err := o(opts); err != nil {
			return nil, errwrap.Wrapf("invalid option: {{err}}", err)
		}
	}
	return opts, nil
}

// With appends middleware to the options.
func With(m ...Handler) Option {
	return func(o *Options) (err error) {
		for _, mid := range m {
			if mid == nil {
				return errors.New("middleware is nil")
			}
		}
		o.Middleware = append(o.Middleware, m...)
		return
	}
}

// With appends middleware funcs to the options.
func WithFunc(mf ...HandlerFunc) Option {
	return func(o *Options) (err error) {
		for _, fn := range mf {
			if fn == nil {
				return errors.New("middleware is nil")
			}
			o.Middleware = append(o.Middleware, fn)
		}
		return
	}
}

// UseHandler sets the destination http.Handler to use. This will be the final call that implements
// the logic of the handler.
func UseHandler(h http.Handler) Option {
	return func(o *Options) (err error) {
		if h == nil {
			err = errors.New("handler is nil")
			return
		}
		o.Handler = h
		return
	}
}

// UseHandler wraps UseHandler allowing for convience http.HandlerFunc.
func UseHandlerFunc(h http.HandlerFunc) Option {
	return UseHandler(h)
}

// ForEach iterates over every element in the iterator, invoking the function argument.
func (hs HandlerIterator) ForEach(f func(Handler)) {
	for _, h := range hs {
		f(h)
	}
}

// Count returns the number of elements in the HandlerIterator.
func (hs HandlerIterator) Count() int {
	return len(hs)
}
