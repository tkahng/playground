package repository

import (
	"reflect"
	"testing"
)

func Test_splitTagValueOptions(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		tag     string
		want    *FieldTag
		wantErr bool
	}{
		{
			name: "split regular comma separated value",
			tag:  `db:"notifications,quoted"`,
			want: &FieldTag{
				Tag:   `db`,
				Value: "notifications",
				Options: []*TagOption{
					{
						Key:   "quoted",
						Value: "true",
					},
				},
			},
		},
		{
			name: "many options",
			tag:  `db:"notifications,schema=public,quoted,table=users"`,
			want: &FieldTag{
				Tag:   `db`,
				Value: "notifications",
				Options: []*TagOption{
					{
						Key:   "schema",
						Value: "public",
					},
					{
						Key:   "quoted",
						Value: "true",
					},
					{
						Key:   "table",
						Value: "users",
					},
				},
			},
		},
		{
			name: "split comma with kv option",
			tag:  `db:"notifications,quoted=true"`,
			want: &FieldTag{
				Tag:   `db`,
				Value: "notifications",
				Options: []*TagOption{
					{
						Key:   "quoted",
						Value: "true",
					},
				},
			},
		},
		{
			name: "split no options",
			tag:  `db:"notifications"`,
			want: &FieldTag{
				Tag:   "db",
				Value: "notifications",
			},
		},
		{
			name: "split comma only",
			tag:  `db:","`,
			want: &FieldTag{
				Tag: "db",
			},
		},
		{
			name: "split comma with option only",
			tag:  `db",hello"`,
			want: &FieldTag{
				Tag:     "db",
				Options: []*TagOption{{Key: "hello", Value: "true"}},
			},
		},
		{
			name: "empty",
			tag:  `db:""`,
			// want: &FieldTag{
			// 	Key: "db",
			// },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := SplitTagValueOptions(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("error splitting tag value: %v", err)
			}
			if !reflect.DeepEqual(tag, tt.want) {
				t.Errorf("deep equal failed. got %v, want %v", tag, tt.want)
			}
		})
	}
}

func TestParseTagOption(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s    string
		want *TagOption
	}{
		struct {
			name string
			s    string
			want *TagOption
		}{
			name: "boolean option",
			s:    "somevalue",
			want: &TagOption{
				Key:   "somevalue",
				Value: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTagOption(tt.s)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("ParseTagOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
