package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
)

//go:generate mockgen -destination ./mocks/http_handler.go -mock_names Handler=MockHTTPHandler -package mocks net/http Handler

//go:generate mockgen -destination ./mocks/middleware_handler.go -package mocks github.com/darren-west/middleware Handler
type (
	Runner struct {
		options *Options
	}

	Handler interface {
		ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler)
	}

	HandlerFunc func(w http.ResponseWriter, r *http.Request, next http.Handler)

	HandlerIterator []Handler

	Options struct {
		Middleware HandlerIterator
		Handler    http.Handler
	}

	Option func(*Options) error
)

func (h *Runner) Options() Options {
	return *h.options
}

func (h *Runner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next := h.options.Handler
	for i := h.options.Middleware.Count() - 1; i >= 0; i-- {
		mid := h.options.Middleware[i]
		next = func(n http.Handler) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				mid.ServeHTTP(w, r, n)
			}
		}(next)
	}
	next.ServeHTTP(w, r)
}

func New(setters ...Option) (*Runner, error) {
	opts, err := newOptions(setters...)
	if err != nil {
		return nil, err
	}
	return &Runner{
		options: opts,
	}, nil
}

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	f(w, r, next)
}

func newOptions(optsetters ...Option) (*Options, error) {
	opts := &Options{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprintf(w, "Hello World!")
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

func (hs HandlerIterator) ForEach(f func(Handler)) {
	for _, h := range hs {
		f(h)
	}
}

func (hs HandlerIterator) Count() int {
	return len(hs)
}
