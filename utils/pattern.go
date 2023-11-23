package utils

import (
	"fmt"
	"strings"
)

func ReplacePattern(in string, replaces ...string) (string, error) {
	if strings.Index(in, "/") != -1 {
		return in, nil
	}
	s := strings.Split(in, ".")
	if len(s) != 4 {
		return "", fmt.Errorf("%s need 4 sections", in)
	}
	for i := range s {
		if s[i] != "*" {
			continue
		}
		if len(replaces) <= i || len(replaces[i]) == 0 {
			return "", fmt.Errorf("%s's %d section invalid", in, i)
		}
		s[i] = replaces[i]
	}
	out := s[0]
	for i := 1; i < len(s); i++ {
		out += "." + s[i]
	}

	return out, nil
}
