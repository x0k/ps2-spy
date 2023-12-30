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
	concat(printer extendablePrinter) extendablePrinter
	extend(printers []extendablePrinter) extendablePrinter
	setSeparator(separator string) extendablePrinter
}

type field struct {
	name      string
	separator string
	value     extendablePrinter
}

func newField(name, separator string, value extendablePrinter) field {
	return field{
		name:      name,
		separator: separator,
		value:     value,
	}
}
func (f field) print(writer io.StringWriter) {
	writer.WriteString(f.name)
	writer.WriteString(f.separator)
	f.value.print(writer)
}
func (f field) concat(value extendablePrinter) extendablePrinter {
	f.value = f.value.concat(value)
	return f
}
func (f field) extend(value []extendablePrinter) extendablePrinter {
	f.value = f.value.extend(value)
	return f
}
func (f field) setSeparator(separator string) extendablePrinter {
	f.separator = separator
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
func (b Bool) concat(p extendablePrinter) extendablePrinter { return p }
func (b Bool) extend(ps []extendablePrinter) extendablePrinter {
	return ps[len(ps)-1]
}
func (b Bool) setSeparator(separator string) extendablePrinter { return b }

type Bit bool

func (b Bit) print(writer io.StringWriter) {
	if b {
		writer.WriteString("1")
	} else {
		writer.WriteString("0")
	}
}
func (b Bit) concat(p extendablePrinter) extendablePrinter { return p }
func (b Bit) extend(ps []extendablePrinter) extendablePrinter {
	return ps[len(ps)-1]
}
func (b Bit) setSeparator(separator string) extendablePrinter { return b }

type Int int

func (i Int) print(writer io.StringWriter) {
	writer.WriteString(strconv.Itoa(int(i)))
}
func (i Int) concat(p extendablePrinter) extendablePrinter { return p }
func (i Int) extend(ps []extendablePrinter) extendablePrinter {
	return ps[len(ps)-1]
}
func (i Int) setSeparator(separator string) extendablePrinter { return i }

type Str string

func (s Str) print(writer io.StringWriter) {
	writer.WriteString(string(s))
}
func (s Str) concat(p extendablePrinter) extendablePrinter { return p }
func (s Str) extend(ps []extendablePrinter) extendablePrinter {
	return ps[len(ps)-1]
}
func (s Str) setSeparator(separator string) extendablePrinter { return s }

type List struct {
	values    []extendablePrinter
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
func (l List) concat(value extendablePrinter) extendablePrinter {
	l.values = append(l.values, value)
	return l
}
func (l List) extend(printers []extendablePrinter) extendablePrinter {
	l.values = append(l.values, printers...)
	return l
}
func (l List) setSeparator(separator string) extendablePrinter {
	l.separator = separator
	return l
}

func stringsToList(strings []string) []extendablePrinter {
	values := make([]extendablePrinter, len(strings))
	for i, field := range strings {
		values[i] = Str(field)
	}
	return values
}

type fields struct {
	firstFieldsSeparator string
	fieldsSeparator      string
	keyValueSeparator    string
	subElementsSeparator string
	names                []string
	fields               []extendablePrinter
	count                int
}

func newFields(
	names []string,
	firstSeparator, fieldsSeparator, keyValueSeparator, subElementsSeparator string,
) fields {
	return fields{
		names:                names,
		fields:               make([]extendablePrinter, len(names)),
		firstFieldsSeparator: firstSeparator,
		fieldsSeparator:      fieldsSeparator,
		keyValueSeparator:    keyValueSeparator,
		subElementsSeparator: subElementsSeparator,
	}
}

func (f fields) setRawField(id int, value extendablePrinter) fields {
	if f.fields[id] == nil {
		f.count++
		f.fields[id] = value
	} else {
		f.fields[id] = f.fields[id].concat(value)
	}
	return f
}

func (f fields) setField(id int, value extendablePrinter) fields {
	return f.setRawField(id, field{
		name:      f.names[id],
		separator: f.keyValueSeparator,
		value:     value,
	})
}

func (f fields) concatField(id int, value extendablePrinter) fields {
	if f.fields[id] == nil {
		f.count++
		f.fields[id] = field{
			name:      f.names[id],
			separator: f.keyValueSeparator,
			value: List{
				values:    []extendablePrinter{value},
				separator: f.subElementsSeparator,
			},
		}
	} else {
		f.fields[id] = f.fields[id].concat(value)
	}
	return f
}

func (f fields) extendField(id int, values []extendablePrinter) fields {
	if f.fields[id] == nil {
		f.count++
		f.fields[id] = field{
			name:      f.names[id],
			separator: f.keyValueSeparator,
			value: List{
				values:    values,
				separator: f.subElementsSeparator,
			},
		}
	} else {
		f.fields[id] = f.fields[id].extend(values)
	}
	return f
}

func (f fields) print(writer io.StringWriter) {
	if f.count == 0 {
		return
	}
	writer.WriteString(f.firstFieldsSeparator)
	i := 0
	// No boundary check cause previous check guarantees that
	// here is at least one non-nil field
	for ; f.fields[i] == nil; i++ {
	}
	f.fields[i].print(writer)
	for i++; i < len(f.fields); i++ {
		if f.fields[i] == nil {
			continue
		}
		writer.WriteString(f.fieldsSeparator)
		f.fields[i].print(writer)
	}
}
