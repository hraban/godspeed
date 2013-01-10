// Utility wrappers for HTTP handlers.
package godspeed

import (
	"mime"
	"net/http"
	"regexp"
)

var findext = regexp.MustCompile(`\.\w+$`)

type postHeaderWrapper struct {
	writer http.ResponseWriter
	// Called after the wrapper handler has done its Head business
	posthandler func(http.ResponseWriter)
	written     bool
}

func (w *postHeaderWrapper) Header() http.Header {
	return w.writer.Header()
}

func (w *postHeaderWrapper) Write(data []byte) (int, error) {
	if !w.written {
		w.posthandler(w.writer)
		w.written = true
	}
	return w.writer.Write(data)
}

func (w *postHeaderWrapper) WriteHeader(s int) {
	w.writer.WriteHeader(s)
}

func wrapPostHeader(w http.ResponseWriter, hook func(http.ResponseWriter)) *postHeaderWrapper {
	return &postHeaderWrapper{
		writer:      w,
		posthandler: hook,
	}
}

// Best-effort guessing of mime-type based on extension of request path. Does
// not override content-type if already set.
func Mimetype(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := func(w http.ResponseWriter) {
			if head := w.Header(); head.Get("Content-Type") == "" {
				ext := findext.FindString(r.URL.Path)
				if ctype := mime.TypeByExtension(ext); ctype != "" {
					head.Set("Content-Type", ctype)
				}
			}
		}
		h.ServeHTTP(wrapPostHeader(w, f), r)
	})
}
