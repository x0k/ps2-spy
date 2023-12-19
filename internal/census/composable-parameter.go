package census

import (
	"fmt"
	"reflect"
	"strings"
)

func writeCensusNestedComposableParameter(builder *strings.Builder, op censusNestedComposableParameter) {
	builder.WriteString(op.getField())
	writeCensusComposableParameter(builder, op)
	for i := 0; i < op.getNestedParametersCount(); i++ {
		builder.WriteString("(")
		op.getNestedParameter(i).String(builder)
		builder.WriteString(")")
	}
}

func writeCensusComposableParameter(builder *strings.Builder, op censusComposableParameter) int {
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
			op.writeProperty(builder, tag[:strings.Index(tag, ",")], fieldValue, count)
			count++
		}
	}
	return count
}

func writeCensusComposableParameterValue(builder *strings.Builder, value reflect.Value, spacer string) {
	vi := value.Interface()
	rt := reflect.ValueOf(vi).Kind()
	if rt == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			builder.WriteString(fmt.Sprintf("%s%v", spacer, value.Index(i)))
			if i < value.Len()-1 {
				builder.WriteString(spacer)
			}
		}
	}
	builder.WriteString(fmt.Sprintf("%v", value))
}

func isValueNilOrDefault(value reflect.Value, valType reflect.Type) bool {
	vi := value.Interface()
	switch reflect.TypeOf(vi).Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Slice:
		return value.Len() == 0
	case reflect.Bool:
		return value.Bool() == false
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
