// Copyright © 2013 Hraban Luyat <hraban@0brg.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package godspeed

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

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

// Compress response if possible
func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var closer io.Closer
		f := func(w http.ResponseWriter) io.Writer {
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
				var wr io.WriteCloser
				switch c {
				case "deflate":
					wr = zlib.NewWriter(w)
				case "gzip":
					wr = gzip.NewWriter(w)
				default:
					continue
				}
				head.Set("Content-Encoding", c)
				closer = wr
				return wr
			}
			return w
		}
		h.ServeHTTP(wrapBody(w, f), r)
		if closer != nil {
			// TODO: Error?
			closer.Close()
		}
	})
}
