package census2

import (
	"io"
	"strconv"
)

type printer interface {
	print(writer io.StringWriter)
}

type field[T printer] struct {
	name      string
	separator string
	value     T
}

func (f field[T]) print(writer io.StringWriter) {
	writer.WriteString(f.name)
	writer.WriteString(f.separator)
	f.value.print(writer)
}

type printableBool bool

func (b printableBool) print(writer io.StringWriter) {
	if b {
		writer.WriteString("true")
	} else {
		writer.WriteString("false")
	}
}

type printableInt int

func (i printableInt) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(i)))
}

type printableString string

func (s printableString) print(writer io.StringWriter) {
	writer.WriteString(string(s))
}
