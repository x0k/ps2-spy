package census2

import (
	"io"
	"strconv"
)

type printer interface {
	print(writer io.StringWriter)
}

type extendablePrinter interface {
	printer
	extend(printers []printer) extendablePrinter
}

type field struct {
	name      string
	separator string
	value     extendablePrinter
}

func (f field) print(writer io.StringWriter) {
	writer.WriteString(f.name)
	writer.WriteString(f.separator)
	f.value.print(writer)
}

func (f field) extend(value []printer) extendablePrinter {
	f.value = f.value.extend(value)
	return f
}

type boolField bool

func (b boolField) print(writer io.StringWriter) {
	if b {
		writer.WriteString("true")
	} else {
		writer.WriteString("false")
	}
}
func (b boolField) extend(printers []printer) extendablePrinter { return b }

type intField int

func (i intField) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(i)))
}
func (i intField) extend(printers []printer) extendablePrinter { return i }

type printableString string

func (s printableString) print(writer io.StringWriter) {
	writer.WriteString(string(s))
}
func (s printableString) extend(printers []printer) extendablePrinter { return s }

type list struct {
	values    []printer
	separator string
}

func (l list) print(writer io.StringWriter) {
	if len(l.values) == 0 {
		return
	}
	l.values[0].print(writer)
	for i := 1; i < len(l.values); i++ {
		writer.WriteString(l.separator)
		l.values[i].print(writer)
	}
}

func (l list) extend(printers []printer) extendablePrinter {
	l.values = append(l.values, printers...)
	return l
}

func stringsToList(strings []string) []printer {
	values := make([]printer, len(strings))
	for i, field := range strings {
		values[i] = printableString(field)
	}
	return values
}
