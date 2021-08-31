package chu

import (
	"errors"
	"reflect"
	"regexp"
)

const (
	valTag = "validate"
)

// Validate 对结构体进行参数校验，识别 tag 为 validate
// 返回校验结果及 error 信息
func Validate(x interface{}) (bool, error) {
	pass := true
	var err error
	xv := reflect.Indirect(reflect.ValueOf(x))
	xt := xv.Type()
	//fmt.Printf("xv: %v\n", xv)
	//fmt.Printf("xt: %v\n", xt)
	for i := 0; i < xv.NumField(); i++ {
		fv := xv.Field(i)
		//fmt.Printf("fv: %v\n", fv)
		tag := xt.Field(i).Tag.Get(valTag)
		switch fv.Kind() {
		case reflect.Int:
		case reflect.String:
			if pass, err = validateString(fv.String(), tag); !pass {
				return pass, err
			}
		case reflect.Slice, reflect.Array:
		case reflect.Struct:
			inner := fv.Interface()
			inpass, inerr := Validate(inner)
			if !inpass {
				pass, err = inpass, inerr
				return pass, err
			}
		default:
		}
	}
	return pass, err
}

var (
	emailPattern = regexp.MustCompile(`^[\w\_\d]+@\w+(\.\w+)*$`)
	wordPattern  = regexp.MustCompile(`^\w+$`)
)

// validateString 校验 string 类型的字段，返回校验结果和 error 信息
func validateString(input, tag string) (bool, error) {
	switch tag {
	case "email":
		if !emailPattern.MatchString(input) {
			return false, errors.New("invalid email")
		}
	case "word":
		if !wordPattern.MatchString(input) {
			return false, errors.New("invalid word")
		}
	case "required":
		if len(input) == 0 {
			return false, errors.New("field should no be empty")
		}
	default:
		return true, nil
	}
	return true, nil
}
