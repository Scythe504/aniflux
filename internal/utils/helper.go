package utils

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

func Paginate[T any](data []T, page, perPage int) []T {
	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(data) {
		return []T{}
	}

	if end > len(data) {
		end = len(data)
	}

	return data[start:end]
}

func LogHandlerError(r *http.Request, handler string, err error, fields map[string]any) {
	var b strings.Builder
	fmt.Fprintf(&b, "handler=%s method=%s path=%s", handler, r.Method, r.URL.Path)
	if r.URL.RawQuery != "" {
		fmt.Fprintf(&b, " query=%q", r.URL.RawQuery)
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(&b, " %s=%v", key, fields[key])
	}

	log.Printf("%s error=%v", b.String(), err)
}
