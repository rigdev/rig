package field

import (
	"fmt"
	"regexp"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
)

type Operation string

const (
	AddedOperation    Operation = "ADDED"
	RemovedOperation  Operation = "REMOVED"
	ModifiedOperation Operation = "MODIFIED"
)

type Change struct {
	FieldPath string
	FieldID   string
	From      Value
	To        Value
	Operation Operation
}

func (c Change) ToProto() *capsule.FieldChange {
	fc := &capsule.FieldChange{
		FieldId:      c.FieldID,
		FieldPath:    c.FieldPath,
		OldValueYaml: c.From.AsString,
		NewValueYaml: c.To.AsString,
		Description:  c.String(),
	}
	switch c.Operation {
	case AddedOperation:
		fc.Operation = capsule.FieldOperation_FIELD_OPERATION_ADDED
	case RemovedOperation:
		fc.Operation = capsule.FieldOperation_FIELD_OPERATION_REMOVED
	case ModifiedOperation:
		fc.Operation = capsule.FieldOperation_FIELD_OPERATION_MODIFIED
	}
	return fc
}

func (c Change) String() string {
	var description string
	switch c.Operation {
	case AddedOperation:
		description = fmt.Sprintf("Added %s", pathDescription(c.FieldPath))
	case RemovedOperation:
		description = fmt.Sprintf("Removed %s", pathDescription(c.FieldPath))
	case ModifiedOperation:
		description = fmt.Sprintf("Changed %s from '%v' to '%v'", pathDescription(c.FieldPath), c.From.AsString, c.To.AsString)
	}

	return description
}

var _namedIndexRegexp = regexp.MustCompile(`^(.*)\[\@(.*)\=(.*)\]$`)

func pathDescription(fieldPath string) string {
	sc := scanner.Scanner{}
	sc.Init(strings.NewReader(fieldPath))
	sc.Error = func(*scanner.Scanner, string) {}
	sc.IsIdentRune = func(r rune, pos int) bool { return unicode.IsLetter(r) || r == '_' || (pos > 0 && unicode.IsDigit(r)) }
	sc.Filename = fieldPath + "\t"

	var result []string

	var suffix string

	for {
		n := sc.Scan()
		if n == -1 {
			break
		}

		switch sc.TokenText() {
		case ".":
		case "$":
		case "[":
			s, err := parseNamed(&sc)
			if err != nil {
				panic(err)
			}

			if len(result) > 0 {
				result[len(result)-1] = strings.TrimSuffix(result[len(result)-1], "s")
			}

			suffix = s
		default:
			result = append(result, sc.TokenText())
		}
	}

	return strings.Join(result, ".") + suffix
}

func parseNamed(sc *scanner.Scanner) (string, error) {
	if sc.Scan() != '@' {
		return "", fmt.Errorf("invalid jsonpath")
	}

	sc.Scan()
	name := sc.TokenText()

	if sc.Scan() != '=' {
		return "", fmt.Errorf("invalid jsonpath")
	}

	value := ""
	for {
		switch sc.Scan() {
		case 0:
			return "", fmt.Errorf("invalid jsonpath")
		case ']':
			return fmt.Sprintf(" (with %s %s)", name, value), nil
		default:
			p := sc.TokenText()
			value += p
		}
	}
}
