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

func writeCensusParameter(builder *strings.Builder, op censusComposableParameter) int {
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

var censusBooleans = map[bool]string{true: "1", false: "0"}

func writeCensusParameterValue(builder *strings.Builder, value reflect.Value, spacer string, valueWriter func(v reflect.Value, builder *strings.Builder)) {
	vi := value.Interface()
	rt := reflect.ValueOf(vi).Kind()
	if rt == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if i > 0 {
				builder.WriteString(spacer)
			}
			valueWriter(value.Index(i), builder)
		}
		return
	}
	valueWriter(value, builder)
}

func censusBasicValueMapper(value reflect.Value, builder *strings.Builder) {
	if value.Type().Implements(reflect.TypeOf((*censusParameter)(nil)).Elem()) {
		value.Interface().(censusParameter).write(builder)
	} else {
		builder.WriteString(fmt.Sprintf("%v", value))
	}
}

func censusValueMapperWithBitBooleans(value reflect.Value, builder *strings.Builder) {
	if value.Kind() == reflect.Bool {
		builder.WriteString(censusBooleans[value.Bool()])
	} else {
		censusBasicValueMapper(value, builder)
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
