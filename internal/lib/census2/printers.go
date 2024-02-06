package census2

import (
	"io"
	"strconv"
)

type printer interface {
	print(writer io.StringWriter) error
}

type optional interface {
	isEmpty() bool
}

type optionalPrinter interface {
	printer
	optional
}

type Ptr[V optionalPrinter] struct {
	value *V
}

func NewPtr[V optionalPrinter](p V) Ptr[V] {
	return Ptr[V]{value: &p}
}

func (p Ptr[V]) isEmpty() bool {
	return p.value == nil || (*p.value).isEmpty()
}

func (p Ptr[V]) print(writer io.StringWriter) error {
	return (*p.value).print(writer)
}

func (p Ptr[V]) Set(v V) {
	*p.value = v
}

type Bool bool

func (b Bool) isEmpty() bool {
	return !bool(b)
}

func (b Bool) print(writer io.StringWriter) error {
	var err error
	if b {
		_, err = writer.WriteString("true")
	} else {
		_, err = writer.WriteString("false")
	}
	return err
}

type Bit bool

func (b Bit) isEmpty() bool {
	return !bool(b)
}

func (b Bit) print(writer io.StringWriter) error {
	var err error
	if b {
		_, err = writer.WriteString("1")
	} else {
		_, err = writer.WriteString("0")
	}
	return err
}

type Int int

func (i Int) isEmpty() bool {
	return false
}

func (i Int) print(writer io.StringWriter) error {
	_, err := writer.WriteString(strconv.Itoa(int(i)))
	return err
}

type Str string

func (s Str) isEmpty() bool {
	return len(s) == 0
}

func (s Str) print(writer io.StringWriter) error {
	_, err := writer.WriteString(string(s))
	return err
}

type BoolWithDefaultTrue bool

func (b BoolWithDefaultTrue) isEmpty() bool {
	return bool(b)
}

func (b BoolWithDefaultTrue) print(writer io.StringWriter) error {
	var err error
	if b {
		_, err = writer.WriteString("true")
	} else {
		_, err = writer.WriteString("false")
	}
	return err
}

type BitWithDefaultTrue bool

func (b BitWithDefaultTrue) isEmpty() bool {
	return bool(b)
}

func (b BitWithDefaultTrue) print(writer io.StringWriter) error {
	var err error
	if b {
		_, err = writer.WriteString("1")
	} else {
		_, err = writer.WriteString("0")
	}
	return err
}

type Uint int

func (u Uint) isEmpty() bool {
	return u < 0
}

func (u Uint) print(writer io.StringWriter) error {
	_, err := writer.WriteString(strconv.Itoa(int(u)))
	return err
}

type Field[T optionalPrinter] struct {
	name      string
	separator string
	value     T
}

func (f Field[T]) isEmpty() bool {
	return f.value.isEmpty()
}

func (f Field[T]) print(writer io.StringWriter) error {
	if _, err := writer.WriteString(f.name); err != nil {
		return err
	}
	if _, err := writer.WriteString(f.separator); err != nil {
		return err
	}
	return f.value.print(writer)
}

type List[T optionalPrinter] struct {
	values    []T
	separator string
}

func NewList[T optionalPrinter](values []T, separator string) List[T] {
	return List[T]{values: values, separator: separator}
}

func (list List[T]) isEmpty() bool {
	return len(list.values) == 0
}

func (l List[T]) print(writer io.StringWriter) error {
	return printList(writer, "", l.separator, l.values)
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

func printList[T optionalPrinter](
	writer io.StringWriter,
	firstFieldSeparator string,
	fieldsSeparator string,
	list []T,
) error {
	l := len(list)
	i := 0
	for i < l && list[i].isEmpty() {
		i++
	}
	if i == l {
		return nil
	}
	if _, err := writer.WriteString(firstFieldSeparator); err != nil {
		return err
	}
	if err := list[i].print(writer); err != nil {
		return err
	}
	for i++; i < l; i++ {
		if list[i].isEmpty() {
			continue
		}
		if _, err := writer.WriteString(fieldsSeparator); err != nil {
			return err
		}
		if err := list[i].print(writer); err != nil {
			return err
		}
	}
	return nil
}

type fakeWriter struct {
	str string
}

func (f *fakeWriter) WriteString(s string) (int, error) {
	f.str += s
	return len(s), nil
}

func printerToString(p printer) (string, error) {
	writer := fakeWriter{}
	err := p.print(&writer)
	return writer.str, err
}
