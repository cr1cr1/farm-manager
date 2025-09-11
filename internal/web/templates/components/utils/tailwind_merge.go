package utils

import "strings"

// TwMerge merges class strings with a simple last-wins join.
// This avoids external dependencies. It does not resolve Tailwind conflicts,
// but is sufficient for our usage here.
func TwMerge(classes ...string) string {
	out := make([]string, 0, len(classes))
	for _, c := range classes {
		if c == "" {
			continue
		}
		out = append(out, c)
	}
	return strings.Join(out, " ")
}
