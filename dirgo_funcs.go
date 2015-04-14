package main

import (
	"regexp"
)

func listable(content []byte) bool {
	rListable := regexp.MustCompile(".*Parent Directory.*|.*Directory listing.*|.*Up To .*|.*Al directorio pri.*")
	if len(rListable.FindString(string(content))) > 0 {
		return true
	} else {
		return false
	}
}
