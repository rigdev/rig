package middleware

import (
	"net/http"
)

type MiddlewareHandlerFunc func(r *http.Request) error

type Middleware interface {
	Wrap(next MiddlewareHandlerFunc) MiddlewareHandlerFunc
}
