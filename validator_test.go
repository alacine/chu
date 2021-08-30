package chu

import "testing"

func Test_validateEmail(t *testing.T) {
	type inner struct {
		C string `validate:"required"`
	}
	type args struct {
		Input string `validate:"email"`
		A     string `validate:"word"`
		B     inner
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{"123abc@gmail.com", "abc_123", inner{C: "--"}}, true},
		{"", args{"123abc.com", "abc_123", inner{C: "--"}}, false},
		{"", args{"123abc@gmail.com", "abc-123", inner{C: "--"}}, false},
		{"", args{"123abc@gmail.com", "abc_123", inner{C: ""}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := Validate(tt.args); got != tt.want {
				t.Errorf("validateEmail() = %v, want %v", got, tt.want)
			} else if !got {
				t.Logf("match error msg: %v\n", err)
			}
		})
	}
}
