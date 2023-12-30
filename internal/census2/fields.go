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
	append(printer printer) extendablePrinter
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

func (f field) append(value printer) extendablePrinter {
	f.value = f.value.append(value)
	return f
}

func (f field) extend(value []printer) extendablePrinter {
	f.value = f.value.extend(value)
	return f
}

type Bool bool

func (b Bool) print(writer io.StringWriter) {
	if b {
		writer.WriteString("true")
	} else {
		writer.WriteString("false")
	}
}
func (b Bool) append(printer printer) extendablePrinter    { return b }
func (b Bool) extend(printers []printer) extendablePrinter { return b }

type Bit bool

func (b Bit) print(writer io.StringWriter) {
	if b {
		writer.WriteString("1")
	} else {
		writer.WriteString("0")
	}
}
func (b Bit) append(printer printer) extendablePrinter    { return b }
func (b Bit) extend(printers []printer) extendablePrinter { return b }

type Int int

func (i Int) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(i)))
}
func (i Int) append(printer printer) extendablePrinter    { return i }
func (i Int) extend(printers []printer) extendablePrinter { return i }

type Str string

func (s Str) print(writer io.StringWriter) {
	writer.WriteString(string(s))
}
func (s Str) append(printer printer) extendablePrinter    { return s }
func (s Str) extend(printers []printer) extendablePrinter { return s }

type List struct {
	values    []printer
	separator string
}

func (l List) print(writer io.StringWriter) {
	if len(l.values) == 0 {
		return
	}
	l.values[0].print(writer)
	for i := 1; i < len(l.values); i++ {
		writer.WriteString(l.separator)
		l.values[i].print(writer)
	}
}

func (l List) append(value printer) extendablePrinter {
	l.values = append(l.values, value)
	return l
}
func (l List) extend(printers []printer) extendablePrinter {
	l.values = append(l.values, printers...)
	return l
}

func stringsToList(strings []string) []printer {
	values := make([]printer, len(strings))
	for i, field := range strings {
		values[i] = Str(field)
	}
	return values
}
