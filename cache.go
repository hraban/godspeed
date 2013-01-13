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
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hraban/lrucache"
	"github.com/hraban/mkdtemp"
)

type cacheWrapper struct {
	Basedir string
	idx     *lrucache.Cache
	wrapped http.Handler
}

type diskEntry struct {
	path string
	size int64
}

func (e diskEntry) Size() int64 {
	return e.size
}

func (e diskEntry) OnPurge(why lrucache.PurgeReason) {
	err := os.Remove(e.path)
	if err != nil {
		// Unexpected and seemingly harmless so I don't really care
		log.Print("Failed to remove", e.path, "from godspeed cache:", err)
	}
}

func cachepath(c *cacheWrapper, r *http.Request) string {
	if r.URL.RawQuery != "" || strings.HasSuffix(r.URL.Path, "/") {
		return ""
	}
	host := r.Host
	if host == "" {
		host = "localhost"
	}
	return c.Basedir + host + "/" + r.URL.Path
}

// Uses X-Cache header to determine cacheability (POC)
// Obvious TODO: real HTTP/1.1 conforming caching
func (c *cacheWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cachepath(c, r)
	if path == "" {
		c.wrapped.ServeHTTP(w, r)
		return
	}
	_, err := c.idx.Get(path)
	if err == nil {
		// Element is cached
		// TODO: Race condition (could be deleted by now)
		// TODO: Breaks (semi transparently) for paths ending with index.html
		w.Header().Set("X-Cache", "Hit")
		http.ServeFile(w, r, path)
		return
	}
	var cachef *os.File
	f := func(w http.ResponseWriter) io.Writer {
		head := w.Header()
		cached := "0"
		defer func() {
			head.Set("X-Cached", cached)
		}()
		if head.Get("X-Cache") == "" {
			return w
		}
		head.Set("X-Cache", "Miss")
		cachedir := dirname(path)
		err := os.MkdirAll(cachedir, 0700)
		if err != nil {
			// Probrem? Just continue as if nothing happened
			log.Printf("Could not create cache file dir %q: %s",
				cachedir, err.Error())
			return w
		}
		cachef, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Could not open cache file %q for writing: %s",
				path, err.Error())
			return w
		}
		cached = "1"
		return io.MultiWriter(cachef, w)
	}
	c.wrapped.ServeHTTP(wrapBody(w, f), r)
	if cachef == nil {
		return
	}
	stat, err := cachef.Stat()
	if err != nil {
		log.Printf("Failed to obtain stat info for %q: %v", path, err)
		cachef.Close()
		os.Remove(path)
		return
	}
	err = cachef.Close()
	if err != nil {
		log.Printf("Could not save cache file %q: %v", path, err)
		// No problem
		return
	}
	c.idx.Set(path, diskEntry{path: path, size: stat.Size()})
	return
}

func mustmkdtemp(tmpl string) string {
	dname, err := mkdtemp.Mkdtemp(tmpl)
	if err != nil {
		panic("Failed to create temporary caching directory: " + err.Error())
	}
	return dname
}

// Store cacheable resources on disk. Naive (and non-conforming) implementation
// of a HTTP cache. Note that this is really not transparent caching:
//
// - Headers are not cached
// - As a consequence, encoded data (gzip etc) is served from cache as raw data
// - Upstream status code is completely ignored
// - Caching directives from client are ignored
// - HTTP caching directives from upstream are ignored
// - Non-standard upstream "X-Cache" header is used to determine cacheability
// - Cache hits are served as raw files, introducing some illegal headers
// - and probably more...
//
// All of that notwithstanding, this is a proof of concept worth exploring.
func Cache(h http.Handler) http.Handler {
	return &cacheWrapper{
		Basedir: mustmkdtemp("godspeed-cache-XXXXXXXX") + "/",
		// TODO: Manage maximum size
		idx:     lrucache.New(1 << 20),
		wrapped: h,
	}
}
