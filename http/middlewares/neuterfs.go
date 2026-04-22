package middlewares

import (
	"net/http"
	"strings"
)

// NeuterFS is an HTTP middleware that prevents directory listing by returning
// 404 Not Found for requests to directory paths (paths ending with "/" or empty paths).
//
// This is a security measure for static file servers to prevent exposing
// directory contents to users.
//
// Example:
//   - /files/ → 404 Not Found (directory listing prevented)
//   - /files/document.pdf → passed through to next handler
func NeuterFS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
