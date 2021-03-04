package glob

import (
	"regexp"
	"strings"
)

var replacer = strings.NewReplacer(
	"*", "[a-zA-Z0-9_]*",
	".", "\\.",
	"?", "[a-zA-Z0-9_]?",
)

func Compile(s string) (*regexp.Regexp, error) {
	s = replacer.Replace(s)
	s = "(?i)^" + s + "$"
	return regexp.Compile(s)
}

func MustCompile(s string) *regexp.Regexp {
	r, err := Compile(s)
	if err != nil {
		panic(err)
	}
	return r
}
