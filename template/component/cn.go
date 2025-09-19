package common

import (
	"regexp"
	"strings"
)

var ws = regexp.MustCompile(`\s+`)

func Cn(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return ws.ReplaceAllString(strings.Join(out, " "), " ")
}
