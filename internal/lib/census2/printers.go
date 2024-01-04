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

type List[T optionalPrinter] struct {
	values    []T
	separator string
}

func (list List[T]) isEmpty() bool {
	return len(list.values) == 0
}

func (l List[T]) print(writer io.StringWriter) {
	printList(writer, "", l.separator, l.values)
}

func stringsToStr(strings []string) []Str {
	values := make([]Str, len(strings))
	for i, field := range strings {
		values[i] = Str(field)
	}
	return values
}

func StrList(strings ...string) List[Str] {
	return List[Str]{values: stringsToStr(strings), separator: ","}
}

func printList[T optionalPrinter](writer io.StringWriter, firstFieldSeparator string, fieldsSeparator string, list []T) {
	l := len(list)
	i := 0
	for i < l && list[i].isEmpty() {
		i++
	}
	if i == l {
		return
	}
	writer.WriteString(firstFieldSeparator)
	list[i].print(writer)
	for i++; i < l; i++ {
		if list[i].isEmpty() {
			continue
		}
		writer.WriteString(fieldsSeparator)
		list[i].print(writer)
	}
}

type fakeWriter struct {
	str string
}

func (f *fakeWriter) WriteString(s string) (int, error) {
	f.str += s
	return len(s), nil
}

func printerToString(p printer) string {
	writer := fakeWriter{}
	p.print(&writer)
	return writer.str
}
