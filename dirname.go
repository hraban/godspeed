package godspeed

import (
	"strings"
)

// Everything up to the last slash (/), or "." if no slash in path
func dirname(path string) string {
	i := strings.LastIndex(path, "/")
	switch i {
	case -1:
		return "."
	case 0:
		return "/"
	}
	path = strings.TrimRight(path[:i], "/")
	if len(path) == 0 {
		return "/"
	}
	return path
}
