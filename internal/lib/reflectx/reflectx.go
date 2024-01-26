package reflectx

import (
	"fmt"
	"reflect"
)

func ToString(value any) (string, error) {
	if v, ok := value.(string); ok {
		return v, nil
	}
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		return val.String(), nil
	}
	return "", fmt.Errorf("cannot convert %T to string", value)
}
