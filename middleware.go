package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
)

type HandlerImpl struct {
	options *Options
}

func (h *HandlerImpl) Options() Options {
	return *h.options
}

func (h *HandlerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

type HandlerFunc func(w http.ResponseWriter, r *http.Request, next http.Handler)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	f(w, r, next)
}

func New(setters ...Option) (*HandlerImpl, error) {
	opts, err := newOptions(setters...)
	if err != nil {
		return nil, err
	}
	return &HandlerImpl{
		options: opts,
	}, nil
}

//go:generate mockgen -destination ./mocks/http_handler.go -mock_names Handler=MockHTTPHandler -package mocks net/http Handler

//go:generate mockgen -destination ./mocks/middleware_handler.go -package mocks github.com/darren-west/middleware Handler
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler)
}

type Options struct {
	Middleware HandlerIterator
	Handler    http.Handler
}

type Option func(*Options) error

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

type HandlerIterator []Handler

func (hs HandlerIterator) ForEach(f func(Handler)) {
	for _, h := range hs {
		f(h)
	}
}

func (hs HandlerIterator) Count() int {
	return len(hs)
}
