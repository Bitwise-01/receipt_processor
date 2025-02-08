package middleware

import "net/http"

// Middleware defines a function to wrap an http.Handler.
type Middleware func(http.Handler) http.Handler
