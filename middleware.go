package web

import (
	"net/http"
)

// MiddlewareFunc defines the function signature for middleware functions.
type MiddlewareFunc func(next http.Handler) http.Handler
