package repository

import (
	"strings"
)

type tagOption struct {
	Key   string
	Value string
}

func parseTagOption(s string) *tagOption {
	if s == "" {
		return nil
	}
	var option tagOption
	for idx, item := range strings.Split(s, "=") {
		if idx == 0 {
			option.Key = item
		} else if idx == 1 {
			option.Value = item
		} else {
			break
		}
	}
	return &option
}

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
