package godspeed

import (
	"fmt"
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
