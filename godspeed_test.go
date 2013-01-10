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
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var simpleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "test")
})

var jsonHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	fmt.Fprint(w, `{"foo": 123}`)
})

func assert200(t *testing.T, r *http.Request, rec *httptest.ResponseRecorder) {
	if rec.Code != 200 {
		t.Fatalf("Unexpected status code for %q: %d, expected: 200",
			r.URL.String(), rec.Code)
	}
}

func testContentType(t *testing.T, r *http.Request, rec *httptest.ResponseRecorder, expected string) {
	ct := rec.Header().Get("content-type")
	if !strings.EqualFold(ct, expected) {
		t.Errorf("Unexpected content-type for %q: %s, expected: %s",
			r.URL.String(), ct, expected)
	}
}

func TestMimetype(t *testing.T) {
	var rec *httptest.ResponseRecorder
	// Test plain text URL without explicit content type
	rec = httptest.NewRecorder()
	// Just assume err == nil
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	Mimetype(simpleHandler).ServeHTTP(rec, r)
	assert200(t, r, rec)
	testContentType(t, r, rec, "text/plain; charset=utf-8")
	// Test plain text URL with explicit JSON content type
	rec = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/test.txt", nil)
	Mimetype(jsonHandler).ServeHTTP(rec, r)
	assert200(t, r, rec)
	testContentType(t, r, rec, "application/json")
}

func TestCompress(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Body = bytes.NewBuffer(nil)
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	r.Header.Add("Accept-Encoding", " gzip ,deflate ")
	Compress(Mimetype(simpleHandler)).ServeHTTP(rec, r)
	assert200(t, r, rec)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("Response not encoded as gzip")
	}
	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test" {
		t.Errorf("Unexpected response decompressed: %v, expected: 'test'", data)
	}
}

// Do not compress data with unknown mime-type
func TestNoCompress(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Body = bytes.NewBuffer(nil)
	r, _ := http.NewRequest("GET", "/test.txt", nil)
	r.Header.Add("Accept-Encoding", "deflate, gzip")
	// Not wrapped in Mimetype, so no content-type header
	Compress(simpleHandler).ServeHTTP(rec, r)
	assert200(t, r, rec)
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Response unexpectedly encoded as %q", enc)
	}
	data, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test" {
		t.Errorf("Unexpected uncompressed response: %v, expected: 'test'", data)
	}
}
