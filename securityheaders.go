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
	"net/http"
)

// Set the HTTP header in the response to the given value if the handler does
// not set it to anything.
func setDefaultHeader(h http.Handler, header, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := func(w http.ResponseWriter) io.Writer {
			head := w.Header()
			// Only if no explicit setting exists
			if head.Get(header) == "" {
				head.Set(header, value)
			}
			return w
		}
		h.ServeHTTP(wrapBody(w, f), r)
	})
}

// Disallow resources from being included in (i)frames on other sites unless
// specified otherwise. This is done through the X-Frame-Options header. If a
// handler does not explicitly set this header, it is set to SAMEORIGIN.
//
// http://tools.ietf.org/html/draft-ietf-websec-x-frame-options-01
func XFrameOptions(h http.Handler) http.Handler {
	return setDefaultHeader(h, "X-Frame-Options", "SAMEORIGIN")
}

// Turn on MSIE8 XSS protection filter.
func XXSSProtection(h http.Handler) http.Handler {
	return setDefaultHeader(h, "X-XSS-Protection", "1; mode=block")
}
