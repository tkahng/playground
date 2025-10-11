package repository

import (
	"strings"
)

// abstraction of a field tag.
//
// given that:
//
//	type Foo struct {
//	    Field1 string `tag1="value", tag2:"value,option1name,option2name=value"`
//	}
//
// this would produced 2 FieldTag instances, one with two options and one without.
//
// `tag:"value"` and `tag:"value,option1,option2=value"`
//
// `db:"notifications,quoted,schema=public"`
//
// ${Key}:"${Value},${Options.0.Key},${Options.1.Key}=${SomeValue}"
type FieldTag struct {
	Tag     string
	Value   string
	Options []*TagOption
}

func (t *FieldTag) GetOptionValue(key string) string {
	for _, option := range t.Options {
		if option.Key == key {
			return option.Value
		}
	}
	return ""
}

// TagOption represents a key=value properties of a field tag.
//
// The contents of the tags themselves are comma separated values,
// where the first value is the tag value, and the rest are options.
//
// `tag:"value,option1,option2=valueA,option3=valueB,option4"`
//
// options are either `option.key=option.value` or just `option`.
type TagOption struct {
	Key   string
	Value string
}

func ParseTagOption(s string) *TagOption {
	if s == "" {
		return nil
	}
	var option TagOption
	for idx, item := range strings.Split(s, "=") {
		switch idx {
		case 0:
			option.Key = item
		case 1:
			option.Value = item
		default:

		}

	}
	return &option
}

func SplitTagValueOptions(inutValue string) (*FieldTag, error) {
	var mainValue string
	var options []*TagOption
	for idx, item := range strings.Split(inutValue, ",") {
		if idx == 0 {
			mainValue = item
		} else {
			options = append(options, ParseTagOption(item))
		}
	}
	return &FieldTag{
		Tag:     mainValue,
		Value:   mainValue,
		Options: options,
	}, nil
}
