package census2

import (
	"io"
	"strconv"
)

type printer interface {
	print(writer io.StringWriter)
}

type optional interface {
	isEmpty() bool
}

type optionalPrinter interface {
	printer
	optional
}

type Bool bool

func (b Bool) isEmpty() bool {
	return b == false
}

func (b Bool) print(writer io.StringWriter) {
	if b {
		writer.WriteString("true")
	} else {
		writer.WriteString("false")
	}
}

type Bit bool

func (b Bit) isEmpty() bool {
	return b == false
}

func (b Bit) print(writer io.StringWriter) {
	if b {
		writer.WriteString("1")
	} else {
		writer.WriteString("0")
	}
}

type Int int

func (i Int) isEmpty() bool {
	return false
}

func (i Int) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(i)))
}

type Str string

func (s Str) isEmpty() bool {
	return len(s) == 0
}

func (s Str) print(writer io.StringWriter) {
	writer.WriteString(string(s))
}

type BoolWithDefaultTrue bool

func (b BoolWithDefaultTrue) isEmpty() bool {
	return b == true
}

func (b BoolWithDefaultTrue) print(writer io.StringWriter) {
	if b {
		writer.WriteString("true")
	} else {
		writer.WriteString("false")
	}
}

type BitWithDefaultTrue bool

func (b BitWithDefaultTrue) isEmpty() bool {
	return b == true
}

func (b BitWithDefaultTrue) print(writer io.StringWriter) {
	if b {
		writer.WriteString("1")
	} else {
		writer.WriteString("0")
	}
}

type Uint int

func (u Uint) isEmpty() bool {
	return u < 0
}

func (u Uint) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(u)))
}

type Field[T optionalPrinter] struct {
	name      string
	separator string
	value     T
}

func (f Field[T]) isEmpty() bool {
	return f.value.isEmpty()
}

func (f Field[T]) print(writer io.StringWriter) {
	writer.WriteString(f.name)
	writer.WriteString(f.separator)
	f.value.print(writer)
}

type List[T printer] struct {
	values    []T
	separator string
}

func (list List[T]) isEmpty() bool {
	return len(list.values) == 0
}

func (l List[T]) print(writer io.StringWriter) {
	l.values[0].print(writer)
	for i := 1; i < len(l.values); i++ {
		writer.WriteString(l.separator)
		l.values[i].print(writer)
	}
}

func stringsToStr(strings []string) []Str {
	values := make([]Str, len(strings))
	for i, field := range strings {
		values[i] = Str(field)
	}
	return values
}

func printFields(writer io.StringWriter, firstFieldSeparator string, fieldsSeparator string, fields ...optionalPrinter) {
	l := len(fields)
	i := 0
	for i < l && fields[i].isEmpty() {
		i++
	}
	if i == l {
		return
	}
	writer.WriteString(firstFieldSeparator)
	fields[i].print(writer)
	for i++; i < l; i++ {
		if fields[i].isEmpty() {
			continue
		}
		writer.WriteString(fieldsSeparator)
		fields[i].print(writer)
	}
}
