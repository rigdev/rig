package common

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

func ToYAMLColored(yamlString string) (string, error) {
	out := yaml.MapSlice{}
	if err := yaml.Unmarshal([]byte(yamlString), &out); err != nil {
		return "", err
	}

	writer := &writer{}
	i := indent{
		unit:     "  ",
		listUnit: "- ",
		level:    0,
		isList:   false,
	}
	if err := write(writer, indentChoice{
		primitive: i,
		complex:   i,
	}, out); err != nil {
		return "", err
	}

	return writer.String(), nil
}

func write(writer *writer, indent indentChoice, object any) error {
	switch v := object.(type) {
	case yaml.MapSlice:
		return writeMap(writer, indent, v)
	case uint:
		return writeNumber(writer, indent.primitive, v)
	case uint8:
		return writeNumber(writer, indent.primitive, v)
	case uint16:
		return writeNumber(writer, indent.primitive, v)
	case uint32:
		return writeNumber(writer, indent.primitive, v)
	case uint64:
		return writeNumber(writer, indent.primitive, v)
	case int:
		return writeNumber(writer, indent.primitive, v)
	case int8:
		return writeNumber(writer, indent.primitive, v)
	case int16:
		return writeNumber(writer, indent.primitive, v)
	case int32:
		return writeNumber(writer, indent.primitive, v)
	case int64:
		return writeNumber(writer, indent.primitive, v)
	case float32:
		return writeNumber(writer, indent.primitive, v)
	case float64:
		return writeNumber(writer, indent.primitive, v)
	case bool:
		return writeBool(writer, indent.primitive, v)
	case string:
		return writeString(writer, indent.primitive, v)
	case nil:
		return writeNull(writer, indent.primitive)
	case []any:
		return writeList(writer, indent.complex, v)
	default:
		return fmt.Errorf("unexpected type: %T", object)
	}
}

func writeMap(writer *writer, indent indentChoice, object yaml.MapSlice) error {
	if len(object) == 0 {
		writer.write(indent.primitive, colorDefault("{}"))
		return nil
	}

	for idx, v := range object {
		if err := writeKeyValue(writer, indent.complex, v.Key.(string), v.Value); err != nil {
			return err
		}
		if idx == 0 {
			indent.complex.isList = false
		}
		if idx < len(object)-1 {
			writer.write(colorNo("\n"))
		}
	}
	return nil
}

func writeNumber(writer *writer, indent indent, object any) error {
	writer.write(indent, colorNumeric(fmt.Sprintf("%v", object)))
	return nil
}

func writeBool(writer *writer, indent indent, object bool) error {
	writer.write(indent, colorNumeric(fmt.Sprintf("%v", object)))
	return nil
}

func writeNull(writer *writer, indent indent) error {
	writer.write(indent, colorDefault("null"))
	return nil
}

func writeKeyValue(writer *writer, ind indent, key string, value any) error {
	suffix := " "
	if !isPrimitive(value) {
		suffix = "\n"
	}
	writer.write(ind, colorKey(key), colorDefault(":"), colorNo(suffix))
	ind.level++
	return write(writer, indentChoice{
		primitive: indent{},
		complex:   ind,
	}, value)
}

func writeList(writer *writer, ind indent, object []any) error {
	ind.level++
	ind.isList = true

	indChoice := indentChoice{}
	elementSuffix := ", "

	primitive := isPrimitive(object)
	if !primitive {
		indChoice.primitive = ind
		indChoice.complex = ind
		elementSuffix = "\n"
	}

	if primitive {
		writer.write(colorDefault("["))
	}
	for idx, v := range object {
		if err := write(writer, indChoice, v); err != nil {
			return err
		}
		if idx < len(object)-1 {
			writer.write(colorDefault(elementSuffix))
		}
	}
	if primitive {
		writer.write(colorDefault("]"))
	}

	return nil
}

func writeString(writer *writer, indent indent, object string) error {
	writer.write(indent, colorString(object))
	return nil
}

type indentChoice struct {
	primitive indent
	complex   indent
}

type indent struct {
	unit     string
	listUnit string
	level    int
	isList   bool
}

func (i indent) String() string {
	if i.isList {
		return colorDefault(strings.Repeat(i.unit, i.level-1) + i.listUnit).String()
	} else {
		return strings.Repeat(i.unit, i.level)
	}
}

func isPrimitive(object any) bool {
	switch v := object.(type) {
	case yaml.MapSlice:
		return len(v) == 0
	case []any:
		for _, vv := range v {
			if !isPrimitive(vv) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

type yamlColor int

const (
	_colorNo = iota
	_colorDefault
	_colorString
	_colorNumeric
	_colorKey
)

func (y yamlColor) String() string {
	switch y {
	case _colorDefault:
		return "[white]"
	case _colorString:
		return "[lime]"
	case _colorNumeric:
		return "[fuchsia]"
	case _colorKey:
		return "[aqua]"
	default:
		return ""
	}
}

func colorNo(s string) colored {
	return newColored(_colorNo, s)
}

func colorDefault(s string) colored {
	return newColored(_colorDefault, s)
}

func colorString(s string) colored {
	return newColored(_colorString, s)
}

func colorNumeric(s string) colored {
	return newColored(_colorNumeric, s)
}

func colorKey(s string) colored {
	return newColored(_colorKey, s)
}

func newColored(c yamlColor, s string) colored {
	return colored{
		color:  c,
		string: s,
	}
}

type colored struct {
	color  yamlColor
	string string
}

func (c colored) String() string {
	return c.color.String() + c.string
}

type writer struct {
	buffer bytes.Buffer
}

func (w *writer) write(values ...fmt.Stringer) {
	for _, v := range values {
		w.buffer.WriteString(v.String())
	}
}

func (w *writer) String() string {
	return w.buffer.String()
}
