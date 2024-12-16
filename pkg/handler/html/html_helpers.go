package html

import "strings"

func Ellipsis(entry string, length int, indicator string) string {
	if entry == "" {
		return ""
	}
	if len(entry) < length {
		return entry
	}
	return entry[:length] + indicator
}

func EnsureTrailingSlash(entry string) string {
	if strings.HasSuffix(entry, "/") {
		return entry
	}
	return entry + "/"
}

func ClassCond(starter, conditional string, condition bool) string {
	classes := make([]string, 1)
	classes = append(classes, starter)
	if condition {
		classes = append(classes, conditional)
	}
	return strings.Join(classes, " ")
}
