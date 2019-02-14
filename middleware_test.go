package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darren-west/middleware"
	"github.com/darren-west/middleware/mocks"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/errwrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_UseHandler(t *testing.T) {
	h := mocks.NewMockHTTPHandler(gomock.NewController(t))
	m, err := middleware.New(
		middleware.UseHandler(h),
	)
	require.NoError(t, err)
	assert.Equal(t, h, m.Options().Handler)
}

func TestOptions_UseHandlerNil(t *testing.T) {
	_, err := middleware.New(
		middleware.UseHandler(nil),
	)
	require.Error(t, err)
	assert.True(t, errwrap.Contains(err, "handler is nil"))
}

func TestOptions_With(t *testing.T) {
	mock := mocks.NewMockHandler(gomock.NewController(t))
	m, err := middleware.New(
		middleware.With(mock),
	)
	require.NoError(t, err)
	require.Equal(t, 1, m.Options().Middleware.Count())
	m.Options().Middleware.ForEach(func(mid middleware.Handler) {
		assert.Equal(t, mock, mid)
	})
}

func TestOptions_WithNil(t *testing.T) {
	_, err := middleware.New(
		middleware.With(nil),
	)
	require.Error(t, err)
	assert.True(t, errwrap.Contains(err, "middleware is nil"))
}

func TestOptions_WithFuncNil(t *testing.T) {
	_, err := middleware.New(
		middleware.WithFunc(nil),
	)
	require.Error(t, err)
	assert.True(t, errwrap.Contains(err, "middleware is nil"))
}

func TestOptions_Default(t *testing.T) {
	m, err := middleware.New()
	require.NoError(t, err)
	assert.NotNil(t, m.Options().Handler)
	assert.Equal(t, 0, m.Options().Middleware.Count())

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://bar", nil)
	m.ServeHTTP(rec, req)
	assert.Equal(t, "hello world!", rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHandlerInvoked(t *testing.T) {
	cont := gomock.NewController(t)
	defer cont.Finish()
	mock := mocks.NewMockHTTPHandler(cont)

	handler, err := middleware.New(middleware.UseHandler(mock))
	require.NoError(t, err)

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
	mock.EXPECT().ServeHTTP(rec, req).
		Return().
		Times(1)
	handler.ServeHTTP(rec, req)
}

func TestHandlerMiddleware_Invoked(t *testing.T) {
	cont := gomock.NewController(t)
	defer cont.Finish()
	mockHandler := mocks.NewMockHTTPHandler(cont)
	mockMiddleware := mocks.NewMockHandler(cont)
	handler, err := middleware.New(
		middleware.With(
			mockMiddleware,
		),
		middleware.UseHandler(mockHandler),
	)
	require.NoError(t, err)

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
	mockHandler.EXPECT().ServeHTTP(rec, req).Return().Times(1)
	mockMiddleware.EXPECT().ServeHTTP(rec, req, gomock.Any()).
		Return().
		Times(1).
		Do(func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
			next(w, r)
		})
	handler.ServeHTTP(rec, req)
}

func TestHandlerMiddleware_NotInvoked(t *testing.T) {
	cont := gomock.NewController(t)
	defer cont.Finish()
	mockHandler := mocks.NewMockHTTPHandler(cont)
	mockMiddleware := mocks.NewMockHandler(cont)
	handler, err := middleware.New(
		middleware.With(
			mockMiddleware,
		),
		middleware.UseHandler(mockHandler),
	)
	require.NoError(t, err)

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
	mockMiddleware.EXPECT().ServeHTTP(rec, req, gomock.Any()).
		Return().
		Times(1)
	handler.ServeHTTP(rec, req)
}

func TestHandlerMiddleware_MultipleMiddleware(t *testing.T) {
	cont := gomock.NewController(t)
	defer cont.Finish()
	mockHandler := mocks.NewMockHTTPHandler(cont)
	handler, err := middleware.New(
		middleware.WithFunc(
			func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
				fmt.Fprintf(w, "1")
				next(w, r)
			},
			func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
				assert.Equal(t, "1", w.(*httptest.ResponseRecorder).Body.String())
				fmt.Fprintf(w, "2")
				next(w, r)
			},
			func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
				assert.Equal(t, "12", w.(*httptest.ResponseRecorder).Body.String())
				fmt.Fprintf(w, "3")
				next(w, r)
			},
			func(w http.ResponseWriter, r *http.Request, next middleware.Next) {
				assert.Equal(t, "123", w.(*httptest.ResponseRecorder).Body.String())
				fmt.Fprintf(w, "4")
				next(w, r)
			},
		),
		middleware.UseHandler(mockHandler),
	)
	require.NoError(t, err)

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/foo", nil)
	mockHandler.EXPECT().ServeHTTP(rec, req).Times(1)

	handler.ServeHTTP(rec, req)
	assert.Equal(t, "1234", rec.Body.String())
}

func TestHandlerInvoked_UseHandlerFunc(t *testing.T) {
	h, err := middleware.New(
		middleware.UseHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "foo")
		}),
	)
	require.NoError(t, err)

	rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://bar", nil)
	h.ServeHTTP(rec, req)
	assert.Equal(t, "foo", rec.Body.String())
}
