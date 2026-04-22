package middlewares

import (
	"mime"
	"net/http"
	"strings"
)

// MimeType is an HTTP middleware that automatically sets the Content-Type header
// based on the file extension in the request URL path.
//
// This is useful for static file servers where the Content-Type might not be set
// correctly by the underlying file system handler.
//
// Example:
//   - /styles/app.css → Content-Type: text/css
//   - /scripts/main.js → Content-Type: application/javascript
//   - /images/logo.png → Content-Type: image/png
func MimeType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileDotExt := r.URL.Path[strings.LastIndex(r.URL.Path, "."):]
		mimeType := mime.TypeByExtension(fileDotExt)

		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
		next.ServeHTTP(w, r)
	})
}
