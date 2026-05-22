package controllers

import (
	"errors"
	"strconv"
	"strings"
)

// split a URL and parse the path parameter at the specified index
func parseURLID(path string, index int) (uint, error) {
	parts := splitURLPath(path)

	if len(parts) <= index {
		return 0, errors.New("index out of path range")
	}

	idVal, err := strconv.ParseUint(parts[index], 10, 32)
	if err != nil {
		return 0, errors.New("invalid identifier parameter format")
	}

	return uint(idVal), nil
}

// split a URL path into its components
func splitURLPath(path string) []string {
	cleanPath := strings.Trim(path, "/")
	return strings.Split(cleanPath, "/")
}
