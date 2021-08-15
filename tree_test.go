package chu

import (
	"reflect"
	"testing"
)

func Test_pathToSeg(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"should pass", args{"/api/"}, []string{"", "api"}},
		{"should pass", args{"/"}, []string{""}},
		{"should pass", args{"/book"}, []string{"", "book"}},
		{"should pass", args{"/book/"}, []string{"", "book"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := pathToSegs(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pathToSeg() = %v, want %v", got, tt.want)
			}
		})
	}
}
