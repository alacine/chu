package chu

import (
	"reflect"
	"testing"
)

func Test_pathToSeg(t *testing.T) {
	type args struct {
		name string
		path string
		want []string
	}
	tests := []args{
		{"should pass", "/api/", []string{"", "api"}},
		{"should pass", "/api/////", []string{"", "api"}},
		{"should pass", "/", []string{""}},
		{"should pass", "/book", []string{"", "book"}},
		{"should pass", "/book/", []string{"", "book"}},
		{"should pass", "/book/我", []string{"", "book", "我"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := pathToSegs(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pathToSeg() = %v, want %v", got, tt.want)
			}
		})
	}

	// 不合法样例
	tests = []args{
		{"should not pass", "//api/abc", []string{}},
		{"should not pass", "/api//abc", []string{}},
		{"should not pass", "api", []string{}},
		{"should not pass", "/api/::id", []string{}},
		{"should not pass", "/api/:id:name", []string{}},
		{"should not pass", "/api/:id:", []string{}},
		{"should not pass", "/api/:/a", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := pathToSegs(tt.path); err == nil {
				t.Errorf("pathToSeg(%v) = %v, want Error", tt.path, got)
			}
		})
	}
}
