package main

import (
	"regexp"
	"strings"
)

var nameRegex = regexp.MustCompile(`[A-z ,.'-]*`)

func cleanName(n string) string {
	matches := nameRegex.FindAllString(n, -1)
	// fmt.Printf("%+v\n", matches)
	cleaned := strings.Join(matches, "")
	return cleaned
}
