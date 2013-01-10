// Utility wrappers for HTTP handlers.
package godspeed

import (
	"mime"
	"net/http"
	"regexp"
)

var findext = regexp.MustCompile(`\.\w+$`)

// Best-effort guessing of mime-type based on extension of request path.
func Mimetype(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := findext.FindString(r.URL.Path)
		if ctype := mime.TypeByExtension(ext); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}
		h.ServeHTTP(w, r)
	})
}
