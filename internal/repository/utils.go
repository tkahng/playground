package repository

import (
	"strings"
)

func splitTagValueOptions(tagValue string) (string, []string) {
	var value string
	var options []string
	for idx, item := range strings.Split(tagValue, ",") {
		if idx == 0 {
			value = item
		} else {
			options = append(options, item)
		}
	}
	return value, options
}
