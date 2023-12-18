package utils

import (
	"reflect"
	"strconv"
)

func SetValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.String:
		field.SetString(value)
	default:
		// 能走到这说明 ScrcpyOptions 结构体有问题
		panic("unsupported type: " + field.Kind().String())
	}
	return nil
}
