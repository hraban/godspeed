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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testHandlerCache = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Cache", "1")
	fmt.Fprintf(w, "test")
})

func TestCache(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Body = bytes.NewBuffer(nil)
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	h := Cache(testHandlerCache)
	h.ServeHTTP(rec, r)
	assert200(t, r, rec)
	// Hack for test purposes
	basedir := h.(*cacheWrapper).Basedir
	cachepath := basedir + "localhost/test.txt"
	data, err := ioutil.ReadFile(cachepath)
	if err != nil {
		t.Fatalf("Failed to open cache file %q: %v", cachepath, err)
	}
	if string(data) != "test" {
		t.Fatalf("Unexpected data in cache file %q: %v, expected: 'test'",
			cachepath, data)
	}
	// Repeat the request, see if it is still correct
	rec = httptest.NewRecorder()
	rec.Body = bytes.NewBuffer(nil)
	r, _ = http.NewRequest("GET", "/test.txt", nil)
	// Reuse handler
	h.ServeHTTP(rec, r)
	assert200(t, r, rec)
	data, _ = ioutil.ReadAll(rec.Body)
	if string(data) != "test" {
		t.Fatalf("Unexpected cached response: %v, expected: 'test'", data)
	}
	os.RemoveAll(basedir)
}
