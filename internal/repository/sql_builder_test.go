package repository

import (
	"testing"
)

func Test_splitTagValueOptions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		tagValue string
		value    string
		options  []string
	}{
		{
			name:     "split regular comma separated value",
			tagValue: "notifications,quoted",
			value:    "notifications",
			options:  []string{"quoted"},
		},
		{
			name:     "split no options",
			tagValue: "notifications",
			value:    "notifications",
			options:  nil,
		},
		{
			name:     "split comma only",
			tagValue: ",",
			value:    "",
			options:  nil,
		},
		{
			name:     "split comma only",
			tagValue: ",hello",
			value:    "",
			options:  []string{"hello"},
		},
		{
			name:     "empty",
			tagValue: "",
			value:    "",
			options:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := splitTagValueOptions(tt.tagValue)
			if got != tt.value {
				t.Errorf("splitDbTag() = %v, want %v", got, tt.value)
			}
			for idx, item := range tt.options {
				if got2[idx] != item {
					t.Errorf("splitDbTag() = %v, want %v", got2[idx], item)
				}
			}
		})
	}
}
