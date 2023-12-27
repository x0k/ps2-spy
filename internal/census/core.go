package census

import (
	"reflect"
	"strings"
)

const VERSION = "0.0.1"

const queryTagName = "queryProp"

type parameter interface {
	write(builder *strings.Builder)
}

type composableParameter interface {
	parameter
	writeProperty(builder *strings.Builder, key string, value reflect.Value, i int)
}

type nestedParameter interface {
	composableParameter
	field() string
	nestedParametersCount() int
	nestedParameter(i int) nestedParameter
}
