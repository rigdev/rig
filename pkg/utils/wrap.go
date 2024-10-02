package utils

import (
	"bytes"
	"strings"
	"unicode"
)

func WordWrap(s string, width int, indent string) string {
	var buffer bytes.Buffer

	lines := strings.Split(s, "\n")

	for _, line := range lines {
		wrapLine(line, width, indent, &buffer)
	}

	return buffer.String()
}

func wrapLine(line string, width int, indent string, buffer *bytes.Buffer) {
	if len(line) == 0 {
		buffer.WriteString("\n")
		return
	}

	var units []unit
	curUnit := unit{
		isWhitespace: unicode.IsSpace(rune(line[0])),
	}
	startIdx := 0
	for idx := 0; idx < len(line); idx++ {
		if unicode.IsSpace(rune(line[idx])) != curUnit.isWhitespace {
			curUnit.s = line[startIdx:idx]
			units = append(units, curUnit)
			curUnit = unit{isWhitespace: !curUnit.isWhitespace}
			startIdx = idx
		}
	}
	curUnit.s = line[startIdx:]
	units = append(units, curUnit)

	curLineLength := 0
	var curLine []unit
	for idx, unit := range units {
		if curLineLength+len(unit.s)+len(indent) > width {
			writeUnits(curLine, width, indent, buffer)
			curLineLength, curLine = 0, nil
		}
		if !unit.isWhitespace || idx == 0 || curLineLength != 0 {
			curLine = append(curLine, unit)
			curLineLength += len(unit.s)
		}
	}
	writeUnits(curLine, width, indent, buffer)
}

func writeUnits(units []unit, width int, indent string, buffer *bytes.Buffer) {
	if units[len(units)-1].isWhitespace {
		units = units[:len(units)-1]
	}
	w := 0
	for _, u := range units {
		w += len(u.s)
	}
	missing := width - w - len(indent)
	if missing < 10 {
		idx := 0
		for missing > 0 {
			if !units[idx].isWhitespace && idx != len(units)-1 {
				units[idx].s += " "
				missing -= 1
			}
			idx = (idx + 1) % len(units)
		}
	}
	buffer.WriteString(indent)
	for _, u := range units {
		buffer.WriteString(u.s)
	}
	buffer.WriteString("\n")
}

type unit struct {
	s            string
	isWhitespace bool
}
