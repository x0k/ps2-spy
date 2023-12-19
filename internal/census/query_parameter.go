package census

import (
	"fmt"
	"reflect"
	"strings"
)

func writeCensusNestedParameter(builder *strings.Builder, op censusNestedParameter) {
	builder.WriteString(op.getField())
	writeCensusParameter(builder, op)
	for i := 0; i < op.getNestedParametersCount(); i++ {
		builder.WriteString("(")
		op.getNestedParameter(i).write(builder)
		builder.WriteString(")")
	}
}

func writeCensusParameter(builder *strings.Builder, op censusParameter) int {
	v := reflect.ValueOf(op)
	ind := reflect.Indirect(v)
	t := ind.Type()
	count := 0
	for i := 0; i < t.NumField(); i++ {
		if tag, ok := t.Field(i).Tag.Lookup(queryTagName); ok {
			fieldValue := ind.Field(i)
			fieldType := fieldValue.Type()
			if isValueNilOrDefault(fieldValue, fieldType) || isValueTagDefault(fieldValue, tag) {
				continue
			}
			key := tag
			index := strings.Index(tag, ",")
			if index > 0 {
				key = tag[:index]
			}
			op.writeProperty(builder, key, fieldValue, count)
			count++
		}
	}
	return count
}

func writeCensusParameterValue(builder *strings.Builder, value reflect.Value, spacer string) {
	vi := value.Interface()
	rt := reflect.ValueOf(vi).Kind()
	if rt == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if i > 0 {
				builder.WriteString(spacer)
			}
			builder.WriteString(fmt.Sprintf("%v", value.Index(i)))
		}
		return
	}
	builder.WriteString(censusReflectValueToString(value, rt))
}

func censusReflectValueToString(v reflect.Value, rt reflect.Kind) string {
	switch rt {
	case reflect.Bool:
		if v.Bool() {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isValueNilOrDefault(value reflect.Value, valType reflect.Type) bool {
	vi := value.Interface()
	switch reflect.TypeOf(vi).Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Slice:
		return value.Len() == 0
		// case reflect.Bool:
		// 	return value.Bool() == false
	}
	return false
}

func isValueTagDefault(value reflect.Value, tag string) bool {
	tagArgs := strings.Split(tag, "default=")
	if len(tagArgs) < 2 {
		return false
	}
	vi := value.Interface()
	return fmt.Sprintf("%v", vi) == tagArgs[1]
}
