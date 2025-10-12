package repository

import (
	"strings"
)

// FieldTag is an abstraction of a field tag.
//
// given that:
//
//	type Foo struct {
//	    Field1 string `tag1="value1", tag2:"value2,option1name,option2name=value2"`
//	}
//
// this would produced 2 FieldTag instances, one with two options and one without.
//
//	var fieldTags = []FieldTag{
//		{
//			Tag:   "tag1",
//			Value: "value1",
//		},
//		{
//			Tag:   "tag2",
//			Value: "value2",
//			Options: []*TagOption{
//				{
//					Key:   "option1name",
//					Value: "true",
//				},
//				{
//					Key:   "option2name",
//					Value: "value2",
//				},
//			},
//		},
//	}
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

// GetOptionValue returns the value of the option with the given key if it exists.
func (t *FieldTag) GetOptionValue(key string) string {
	for _, option := range t.Options {
		if option.Key == key {
			return option.Value
		}
	}
	return ""
}

// ParseFieldTag returns the field tag with the given name if it exists.
// otherwise it returns nil
func ParseFieldTag(tagContent string) *FieldTag {
	var mainValue string
	var options []*TagOption
	for idx, item := range strings.Split(tagContent, ",") {
		cleanedItem := strings.TrimSpace(item)
		if cleanedItem == "" {
			continue
		}
		if idx == 0 {
			mainValue = cleanedItem
		} else {
			if option := ParseTagOption(cleanedItem); option != nil {
				options = append(options, option)
			}
		}
	}
	fieldTag := &FieldTag{
		Value:   mainValue,
		Options: options,
	}
	return fieldTag

}

// TagOption represents a key=value properties of a field tag.
//
// The contents of the tags themselves are comma separated values,
// where the first value is the tag value, and the rest are options.
//
// `tag:"value,option1,option2=valueA,option3=valueB,option4"`
//
// options are either `option.key=option.value` or just `option`, in which case the value is `true`.
type TagOption struct {
	Key   string
	Value string
}

// ParseTagOption returns a TagOption from a string.
func ParseTagOption(s string) *TagOption {
	if s == "" {
		return nil
	}
	var option TagOption
	items := strings.Split(s, "=")
	if len(items) == 1 {
		firstItem := strings.TrimSpace(items[0])
		if firstItem == "" {
			return nil
		}
		option.Key = firstItem
		option.Value = "true"
		return &option
	} else if len(items) == 2 {
		option.Key = strings.TrimSpace(items[0])
		option.Value = strings.TrimSpace(items[1])
		return &option
	} else {
		return nil
	}

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
