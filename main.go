// Utility wrappers for HTTP handlers.
package godspeed

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
)

var findext = regexp.MustCompile(`\.\w+$`)

type responseWrapper func(http.ResponseWriter) http.ResponseWriter

type postHeaderWrapper struct {
	wOrig, wNew http.ResponseWriter
	// Called after the wrapper handler has done its Head business
	posthandler responseWrapper
}

func (w *postHeaderWrapper) Header() http.Header {
	return w.wOrig.Header()
}

func (w *postHeaderWrapper) Write(data []byte) (int, error) {
	if w.wNew == nil {
		w.wNew = w.posthandler(w.wOrig)
	}
	return w.wNew.Write(data)
}

func (w *postHeaderWrapper) WriteHeader(s int) {
	w.wOrig.WriteHeader(s)
}

func (w *postHeaderWrapper) Close() error {
	var err, err2 error
	if c, ok := w.wOrig.(io.Closer); ok {
		err = c.Close()
	}
	if c, ok := w.wNew.(io.Closer); ok {
		err2 = c.Close()
	}
	if err != nil {
		return err
	}
	return err2
}

func wrapPostHeader(w http.ResponseWriter, hook responseWrapper) *postHeaderWrapper {
	return &postHeaderWrapper{
		wOrig:       w,
		posthandler: hook,
	}
}

// Best-effort guessing of mime-type based on extension of request path. Does
// not override content-type if already set.
func Mimetype(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := func(w http.ResponseWriter) http.ResponseWriter {
			if head := w.Header(); head.Get("Content-Type") == "" {
				ext := findext.FindString(r.URL.Path)
				if ctype := mime.TypeByExtension(ext); ctype != "" {
					head.Set("Content-Type", ctype)
				}
			}
			return w
		}
		wrap := wrapPostHeader(w, f)
		h.ServeHTTP(wrap, r)
		// TODO: Error?
		wrap.Close()
	})
}

var compressableTypePrefixes = [...]string{
	"text/",
	"application/json",
	"application/xml+xhtml",
	"application/javascript",
}

func compressable(ctype string) bool {
	for _, t := range compressableTypePrefixes {
		if strings.HasPrefix(ctype, t) {
			return true
		}
	}
	return false
}

type wrappedRW struct {
	w io.WriteCloser
	http.ResponseWriter
}

func (w *wrappedRW) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

func (w *wrappedRW) Close() error {
	return w.w.Close()
}

func wrapResponseWriter(rw http.ResponseWriter, w io.WriteCloser) http.ResponseWriter {
	return &wrappedRW{w, rw}
}

func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := func(w http.ResponseWriter) http.ResponseWriter {
			head := w.Header()
			if !compressable(head.Get("Content-Type")) {
				return w
			}
			// Idempotent
			if head.Get("Content-Encoding") != "" {
				return w
			}
			for _, c := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
				// TODO: qvalue
				c = strings.TrimSpace(c)
				if c != "gzip" {
					continue
				}
				head.Set("Content-Encoding", "gzip")
				return wrapResponseWriter(w, gzip.NewWriter(w))
			}
			return w
		}
		wrap := wrapPostHeader(w, f)
		h.ServeHTTP(wrap, r)
		// TODO: Error?
		wrap.Close()
	})
}
