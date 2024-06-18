package common

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"golang.org/x/exp/slices"
)

const AbortedErrMsg = "prompt aborted"

type GetInputOption = func(*textinput.TextInput)

// TODO What about non-string Selection
type SelectInputOption = func(s *selection.Selection[string])

func ValidateAllOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateAll
}

func BoolValidateOpt(inp *textinput.TextInput) {
	inp.Validate = BoolValidate
}

func ValidateIntOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateInt
}

func ValidateUIntOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateUInt
}

func ValidateIntInRangeOpt(minInclusive, maxInclusive int) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.Validate = ValidateIntInRange(minInclusive, maxInclusive)
	}
}

func ValidateNonEmptyOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateNonEmpty
}

func ValidateAbsPathOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateAbsolutePath
}

func ValidateFilePathOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateFilePath
}

func ValidateEmailOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateEmail
}

func ValidateSystemNameOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateSystemName
}

func ValidateKubernetesNameOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateKubernetesName
}

func ValidateURLOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateURL
}

func ValidateImageOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateImage
}

func ValidatePhoneOpt(inp *textinput.TextInput) {
	inp.Validate = utils.ValidatePhone
}

func ValidatePasswordOpt(inp *textinput.TextInput) {
	inp.Validate = utils.ValidatePassword
}

func ValidateBoolOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateBool
}

func ValidateQuantityOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateQuantity
}

func ValidatePortOpt(inp *textinput.TextInput) {
	inp.Validate = ValidatePort
}

func ValidateURLPathOpt(inp *textinput.TextInput) {
	inp.Validate = utils.ValidateURLPath
}

func ValidateLengthOpt(minLength, maxLength int) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.Validate = ValidateLength(minLength, maxLength)
	}
}

func ValidateAndOpt(validators ...func(string) error) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.Validate = ValidateAnd(validators...)
	}
}

func ValidateAllowEmptyOpt(validator func(string) error) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.Validate = ValidateAllowEmpty(validator)
	}
}

func ValidateCronExpressionOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateCronExpression
}

func ValidateUniqueOpt(values []string, msgAndArgs ...string) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.Validate = ValidateUnique(values, msgAndArgs...)
	}
}

func ValidateDurationOpt(inp *textinput.TextInput) {
	inp.Validate = ValidateDuration
}

func InputDefaultOpt(def string) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.InitialValue = def
	}
}

func SelectEnableFilterOpt(s *selection.Selection[string]) {
	s.Filter = selection.FilterContainsCaseSensitive[string]
}

func SelectFuzzyFilterOpt(s *selection.Selection[string]) {
	s.Filter = func(filter string, choice *selection.Choice[string]) bool {
		return fuzzy.Match(filter, choice.Value)
	}
}

func SelectPageSizeOpt(size int) SelectInputOption {
	return func(s *selection.Selection[string]) {
		s.PageSize = size
	}
}

func SelectExtendTemplateOpt(t template.FuncMap) SelectInputOption {
	return func(s *selection.Selection[string]) {
		s.ExtendedTemplateFuncs = t
	}
}

func SelectTemplateOpt(template string) SelectInputOption {
	return func(s *selection.Selection[string]) {
		s.Template = template
	}
}

func SelectDontShowResultOpt(s *selection.Selection[string]) {
	s.ResultTemplate = ""
}

var inputTemplate = `
	{{- Bold .Prompt }} {{ .Input -}}
	{{- if .ValidationError }}
        {{- Foreground "1" (Bold "✘") }}
        {{- if ge (len (StripCursor .Input)) 3 }}
            {{- printf " %s" (Italic (FormatValidationError .ValidationError)) }}
        {{- end }}
	{{- else }} {{ Foreground "2" (Bold "✔") }}
	{{- end -}}
`

func stripCursor(s string) string {
	// When the cursor blinks away, it adds a space instead
	if len(s) > 0 && s[len(s)-1] == ' ' {
		return s[:len(s)-1]
	}
	// The default cursor of the Input field is exactly these bytes
	// There might be a unicode character equivalent or smth, but I'm not sure
	// This was the easiest way of fixing my issue. Don't judge.
	cursorBytes := []byte{27, 91, 55, 109, 32, 27, 91, 48, 109}
	bs := []byte(s)
	if len(bs) < len(cursorBytes) {
		return s
	}
	if bytes.Equal(bs[len(bs)-len(cursorBytes):], cursorBytes) {
		return string(bs[:len(bs)-len(cursorBytes)])
	}

	return s
}

func formatValidationError(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	badPrefixes := []string{"invalid_argument:", "invalid password;"}
	for {
		found := false
		for _, p := range badPrefixes {
			var b bool
			s = strings.TrimSpace(s)
			s, b = strings.CutPrefix(s, p)
			found = found || b
		}
		if !found {
			break
		}
	}
	return s
}

var templateExtensions = map[string]any{
	// The Input variable to the Input template will get the blinking cursor prepended
	// Thus you need to strip it if you want access to the real input
	"StripCursor":           stripCursor,
	"FormatValidationError": formatValidationError,
}

type Prompter interface {
	Input(label string, opts ...GetInputOption) (string, error)
	Password(label string) (string, error)
	Select(label string, choices []string, opts ...SelectInputOption) (int, string, error)
	Confirm(label string, def bool) (bool, error)
	TableSelect(
		label string,
		choices [][]string,
		columnHeaders []string,
		opts ...SelectInputOption,
	) (int, error)
}

type StandardPrompter struct{}

func (StandardPrompter) Input(label string, opts ...GetInputOption) (string, error) {
	input := textinput.New(label)
	input.Template = inputTemplate
	input.ExtendedTemplateFuncs = templateExtensions
	for _, opt := range opts {
		opt(input)
	}

	s, err := input.RunPrompt()
	if err != nil {
		return "", err
	}
	return s, nil
}

func (StandardPrompter) Password(label string) (string, error) {
	input := textinput.New(label)
	input.Hidden = true
	input.Validate = utils.ValidatePassword
	input.Template = inputTemplate
	input.ExtendedTemplateFuncs = templateExtensions

	pw, err := input.RunPrompt()
	if err != nil {
		return "", err
	}
	return pw, nil
}

func (StandardPrompter) Select(label string, choices []string, opts ...SelectInputOption) (int, string, error) {
	sp := selection.New(label, choices)
	sp.Filter = nil
	sp.PageSize = 5
	for _, opt := range opts {
		opt(sp)
	}
	choice, err := sp.RunPrompt()
	if err != nil {
		return 0, "", err
	}

	return slices.Index(choices, choice), choice, nil
}

func (StandardPrompter) Confirm(label string, def bool) (bool, error) {
	input := textinput.New(label)
	input.Validate = ValidateBool
	input.Template = confirmTemplateY
	if !def {
		input.Template = confirmTemplateN
	}
	result, err := input.RunPrompt()
	if err != nil {
		return false, err
	}
	if result == "" {
		return def, nil
	}

	return parseBool(result)
}

// TODO matias@rig.dev Find a better way instead of duplicate template strings
var (
	confirmTemplateY = `
	{{- Bold .Prompt }} {{- Faint " [Y/n]" }} {{ .Input -}}
	{{- if .ValidationError }} {{ Foreground "1" (Bold "✘") }}
	{{- else }} {{ Foreground "2" (Bold "✔") }}
	{{- end -}}
`
	confirmTemplateN = `
	{{- Bold .Prompt }} {{- Faint " [y/N]" }} {{ .Input -}}
	{{- if .ValidationError }} {{ Foreground "1" (Bold "✘") }}
	{{- else }} {{ Foreground "2" (Bold "✔") }}
	{{- end -}}
`
)

func (p StandardPrompter) TableSelect(
	label string,
	choices [][]string,
	columnHeaders []string,
	opts ...SelectInputOption,
) (int, error) {
	if len(choices) == 0 {
		return 0, fmt.Errorf("'%s' failed; no choices to select from", label)
	}

	// TODO Honestly, this thing with manually creating the table rows and header
	// feels like I'm reinventing the wheel. Maybe find some package to do this for me?
	// I can't just use our table pretty printer as I don't want to print a table,
	// I want a string for each individual row and a couple strings for the table headder
	rows, colLengths, err := formatRows(choices, " | ")
	if err != nil {
		return 0, err
	}

	if len(colLengths) != len(columnHeaders) {
		return 0, fmt.Errorf(
			"number of columns in 'choices' (%v) and 'columnHeaders' (%v) don't agree",
			len(colLengths), len(columnHeaders),
		)
	}

	var header string
	for idx, c := range columnHeaders {
		header += text.AlignCenter.Apply(c, colLengths[idx])
	}
	headerBorder := strings.Repeat("-", len(header))

	opts = append(opts,
		SelectExtendTemplateOpt(map[string]any{
			"header":       func() string { return header },
			"headerBorder": func() string { return headerBorder },
		}),
		SelectTemplateOpt(tableSelectTemplate),
	)
	idx, _, err := p.Select(label, rows, opts...)
	return idx, err
}

var tableSelectTemplate = `
{{- if .Prompt -}}
  {{ Bold .Prompt }}
{{ end -}}
{{ if .IsFiltered }}
  {{- print .FilterPrompt " " .FilterInput }}
{{ end }}
{{ print "  " header }}
{{ println "  " headerBorder }}
{{- range  $i, $choice := .Choices }}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}}
  {{- else -}}
    {{- "  " -}}
  {{- end -}}

  {{- if eq $.SelectedIndex $i }}
   {{- print (Foreground "32" (Bold "▸ ")) (Selected $choice) "\n" }}
  {{- else }}
    {{- print "  " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`

func formatRows(rows [][]string, colDelimiter string) ([]string, []int, error) {
	if len(rows) == 0 {
		return nil, nil, nil
	}

	for _, r := range rows[1:] {
		if len(r) != len(rows[0]) {
			return nil, nil, errors.New("the rows are not all of equal length")
		}
	}

	var colLengths []int
	for cIdx := range rows[0] {
		longest := 0
		for _, row := range rows {
			l := len(row[cIdx])
			if l > longest {
				longest = l
			}
		}
		colLengths = append(colLengths, longest)
	}

	var result []string
	for _, row := range rows {
		var s string
		for cIdx, c := range row {
			s += text.AlignLeft.Apply(c, colLengths[cIdx]) + colDelimiter
		}
		result = append(result, s)
	}

	for idx := range colLengths {
		colLengths[idx] += len(colDelimiter)
	}

	return result, colLengths, nil
}

// TODO Check against error type using errors.Is
func ErrIsAborted(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == AbortedErrMsg
}
