package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

func parseIntArg(raw, name string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func emptyAsDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
