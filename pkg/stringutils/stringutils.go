// Package stringutils provides helper functions for dealing with strings.
package stringutils

import (
	"strings"
)

// Ellipsis truncates a string to fit within maxlen, and appends ellipsis (...).
// For maxlen of 3 and lower, no ellipsis is appended.
func Ellipsis(s string, maxlen int) string {
	r := []rune(s)
	if len(r) <= maxlen {
		return s
	}
	if maxlen <= 3 {
		return string(r[:maxlen])
	}
	return string(r[:maxlen-3]) + "..."
}

// InSlice tests whether a string is contained in a slice of strings or not.
// Comparison is case insensitive
func InSlice(slice []string, s string) bool {
	for _, ss := range slice {
		if strings.ToLower(s) == strings.ToLower(ss) {
			return true
		}
	}
	return false
}
