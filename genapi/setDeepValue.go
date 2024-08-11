package genapi

import (
	"fmt"
	"reflect"
	"strconv"
)

func index(v reflect.Value, idx string) reflect.Value {
	if i, err := strconv.Atoi(idx); err == nil {
		return v.Index(i)
	}
	return v.FieldByName(idx)
}

func set(d interface{}, value interface{}, path ...string) {
	v := reflect.ValueOf(d)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic("d must be a non-nil pointer")
	}
	v = v.Elem() // Dereference to get the actual value

	for _, s := range path {
		switch v.Kind() {
		case reflect.Interface:
			v = v.Elem()
		}
		switch v.Kind() {
		case reflect.Interface:
			v = v.Elem()
		case reflect.Slice:
			fmt.Println("setting")
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				set(elem.Addr().Interface(), value, s) // Recursively set the value in each element
			}
			return
		case reflect.Map:
			key := reflect.ValueOf(s)
			v.SetMapIndex(key, reflect.ValueOf(value))
			return
		default:
			panic(fmt.Sprintf("unsupported kind: %s", v.Kind()))
		}
	}
}
