package headers

import (
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}
func isValidTokenChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) ||
		strings.ContainsRune("!#$%&'*+-.^_`|~", r)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	s := string(data)
	crlfIndex := strings.Index(s, "\r\n")
	if crlfIndex == -1 {
		return 0, false, nil
	}

	line := s[:crlfIndex]
	if line == "" {
		return crlfIndex + 2, true, nil
	}

	trimmed := strings.TrimLeft(line, " \t")
	colonIndex := strings.Index(trimmed, ":")
	if colonIndex == -1 || strings.Contains(trimmed[:colonIndex], " ") {
		return 0, false, fmt.Errorf("invalid header format")
	}

	key := strings.TrimSpace(trimmed[:colonIndex])
	value := strings.TrimSpace(trimmed[colonIndex+1:])

	for _, r := range key {
		if !isValidTokenChar(r) {
			return 0, false, fmt.Errorf("invalid character in header key: %q", r)
		}
	}

	normalizedKey := strings.ToLower(key)

	if existing, ok := h[normalizedKey]; ok {
		h[normalizedKey] = existing + ", " + value
	} else {
		h[normalizedKey] = value
	}

	return crlfIndex + 2, false, nil
}
func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == key {
			return v
		}
	}
	return ""
}
func (h Headers) Set(key, value string) {
	if h == nil {
		return
	}
	normalizedKey := strings.ToLower(strings.TrimSpace(key))
	h[normalizedKey] = strings.TrimSpace(value)
}
func (h Headers) Del(key string) {
	if h == nil {
		return
	}
	normalizedKey := strings.ToLower(strings.TrimSpace(key))
	delete(h, normalizedKey)
}

//go test -v ./internal/headers
