package main

import (
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(`[a-z ,.'-]+`)

func cleanName(n string) string {
	return strings.Join(nameRegex.FindAllString(n, 12), "")
}
