package utils

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"
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

// CompareNatural compares two strings naturally (e.g. "S1" < "S2" < "S11").
func CompareNatural(a, b string) bool {
	chunksA := splitChunks(a)
	chunksB := splitChunks(b)

	for i := 0; i < len(chunksA) && i < len(chunksB); i++ {
		cA := chunksA[i]
		cB := chunksB[i]

		if cA == cB {
			continue
		}

		// Try to parse both as integers
		nA, errA := strconv.Atoi(cA)
		nB, errB := strconv.Atoi(cB)

		if errA == nil && errB == nil {
			return nA < nB
		}

		// Otherwise, compare as strings
		return cA < cB
	}

	return len(chunksA) < len(chunksB)
}

func splitChunks(s string) []string {
	var chunks []string
	if len(s) == 0 {
		return chunks
	}

	runes := []rune(s)
	start := 0
	isDigit := unicode.IsDigit(runes[0])

	for i := 1; i < len(runes); i++ {
		currIsDigit := unicode.IsDigit(runes[i])
		if currIsDigit != isDigit {
			chunks = append(chunks, string(runes[start:i]))
			start = i
			isDigit = currIsDigit
		}
	}
	chunks = append(chunks, string(runes[start:]))
	return chunks
}

