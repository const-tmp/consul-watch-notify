package utils

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

var FuncMap = template.FuncMap{
	"join": join,
}

func join(items reflect.Value, sep string) (string, error) {
	var strItems []string

	switch items.Kind() {
	case reflect.Slice, reflect.Array:
		length := items.Len()
		strItems = make([]string, 0, length)
		for i := 0; i < length; i++ {
			strItems = append(strItems, fmt.Sprint(items.Index(i)))
		}

	case reflect.Map:
		keys := items.MapKeys()
		strItems = make([]string, 0, len(keys))
		for _, key := range keys {
			strItems = append(strItems, fmt.Sprint(key.Interface()))
		}

	case reflect.Chan:
		for {
			v, ok := items.Recv()
			if !ok {
				break
			}
			strItems = append(strItems, fmt.Sprint(v.Interface()))
		}
	default:
		return "", fmt.Errorf("%T cannot be joined", items)
	}
	return strings.Join(strItems, sep), nil
}
