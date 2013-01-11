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
	"io"
	"mime"
	"net/http"
	"regexp"
)

var findext = regexp.MustCompile(`\.\w+$`)

// Best-effort guessing of mime-type based on extension of request path. Does
// not override content-type if already set.
func Mimetype(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := func(w http.ResponseWriter) io.Writer {
			if head := w.Header(); head.Get("Content-Type") == "" {
				ext := findext.FindString(r.URL.Path)
				if ctype := mime.TypeByExtension(ext); ctype != "" {
					head.Set("Content-Type", ctype)
				}
			}
			return w
		}
		h.ServeHTTP(wrapBody(w, f), r)
	})
}
