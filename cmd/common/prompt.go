package common

import (
	"errors"
	"html/template"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rigdev/rig/pkg/utils"
	"golang.org/x/exp/slices"
)

type GetInputOption = func(*textinput.TextInput)

// TODO What about non-string Selection
type SelectInputOption = func(s *selection.Selection[string])

var ValidateAllOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateAll
}

var BoolValidateOpt = func(inp *textinput.TextInput) {
	inp.Validate = BoolValidate
}

var ValidateIntOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateInt
}

var ValidateNonEmptyOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateNonEmpty
}

var ValidateAbsPathOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateAbsolutePath
}

var ValidateEmailOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateEmail
}

var ValidateSystemNameOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateSystemName
}

var ValidateURLOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateURL
}

var ValidateImageOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateImage
}

var ValidatePhoneOpt = func(inp *textinput.TextInput) {
	inp.Validate = utils.ValidatePhone
}

var ValidatePasswordOpt = func(inp *textinput.TextInput) {
	inp.Validate = utils.ValidatePassword
}

var ValidateBoolOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateBool
}

var ValidateQuantityOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateQuantity
}

var InputDefaultOpt = func(def string) GetInputOption {
	return func(inp *textinput.TextInput) {
		inp.InitialValue = def
	}
}

var SelectEnableFilterOpt = func(s *selection.Selection[string]) {
	s.Filter = selection.FilterContainsCaseSensitive[string]
}

var SelectFuzzyFilterOpt = func(s *selection.Selection[string]) {
	s.Filter = func(filter string, choice *selection.Choice[string]) bool {
		return fuzzy.Match(filter, choice.Value)
	}
}

var SelectExtendTemplateOpt = func(t template.FuncMap) SelectInputOption {
	return func(s *selection.Selection[string]) {
		s.ExtendedTemplateFuncs = t
	}
}

var SelectTemplateOpt = func(template string) SelectInputOption {
	return func(s *selection.Selection[string]) {
		s.Template = template
	}
}

func PromptInput(label string, opts ...GetInputOption) (string, error) {
	input := textinput.New(label)
	for _, opt := range opts {
		opt(input)
	}

	s, err := input.RunPrompt()
	if err != nil {
		return "", err
	}
	return s, nil
}

func PromptPassword(label string) (string, error) {
	input := textinput.New(label)
	input.Hidden = true
	input.Validate = utils.ValidatePassword
	input.ResultTemplate = ""

	pw, err := input.RunPrompt()
	if err != nil {
		return "", err
	}
	return pw, nil
}

func PromptSelect(label string, choices []string, opts ...SelectInputOption) (int, string, error) {
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

func PromptConfirm(label string, def bool) (bool, error) {
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

func PromptTableSelect(label string, choices [][]string, columnHeaders []string, opts ...SelectInputOption) (int, error) {
	// TODO Honestly, this thing with manually creating the table rows and header
	// feels like I'm reinventing the wheel. Maybe find some package to do this for me?
	// I can't just use our table pretty printer as I don't want to print a table,
	// I want a string for each individual row and a couple strings for the table headder
	rows, colLengths, err := formatRows(choices, " | ")
	if err != nil {
		return 0, err
	}

	if len(colLengths) != len(columnHeaders) {
		return 0, errors.New("number of columns in 'choices' and 'columnHeaders' don't agree")
	}

	var header string
	for idx, c := range columnHeaders {
		header += text.AlignCenter.Apply(c, colLengths[idx])
	}
	headerBorder := strings.Repeat("-", len(header))

	opts = append(opts, SelectExtendTemplateOpt(map[string]any{
		"header":       func() string { return header },
		"headerBorder": func() string { return headerBorder },
	}))
	idx, _, err := PromptSelect(label, rows, opts...)
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
