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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testHandlerXFrameOptions = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Frame-Options", "DENY")
	fmt.Fprint(w, "test")
})

func TestXFrameOptionsDefault(t *testing.T) {
	rec := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	h := XFrameOptions(testHandlerSimple)
	h.ServeHTTP(rec, r)
	assert200(t, r, rec)
	if opt := rec.Header().Get("X-Frame-Options"); opt != "SAMEORIGIN" {
		t.Errorf("Unexpected X-Frame-Options value: %q, expected: SAMEORIGIN",
			opt)
	}
}

func TestXFrameOptionsExplicit(t *testing.T) {
	rec := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	h := XFrameOptions(testHandlerXFrameOptions)
	h.ServeHTTP(rec, r)
	assert200(t, r, rec)
	if opt := rec.Header().Get("X-Frame-Options"); opt != "DENY" {
		t.Errorf("Unexpected X-Frame-Options value: %q, expected: DENY",
			opt)
	}
}
