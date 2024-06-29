package util

import "strings"

func ErrorContains(err error, msg string) bool {
	return strings.Contains(err.Error(), msg)
}

func ErrorNotFound(err error) bool {
	return strings.Contains(err.Error(), "no rows in result set")
}
