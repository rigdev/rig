package common

import (
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rigdev/rig/pkg/utils"
	"golang.org/x/exp/slices"
)

type GetInputOption = func(*textinput.TextInput) *textinput.TextInput

var ValidateAllOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateAll
	return inp
}

var BoolValidateOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = BoolValidate
	return inp
}

var ValidateIntOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateInt
	return inp
}

var ValidateNonEmptyOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateNonEmpty
	return inp
}

var ValidateEmailOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateEmail
	return inp
}

var ValidateSystemNameOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateSystemName
	return inp
}

var ValidateURLOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateURL
	return inp
}

var ValidateImageOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateImage
	return inp
}

var ValidatePhoneOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = utils.ValidatePhone
	return inp
}

var ValidatePasswordOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = utils.ValidatePassword
	return inp
}

var ValidateBoolOpt = func(inp *textinput.TextInput) *textinput.TextInput {
	inp.Validate = ValidateBool
	return inp
}

var InputDefaultOpt = func(def string) GetInputOption {
	return func(inp *textinput.TextInput) *textinput.TextInput {
		inp.InitialValue = def
		return inp
	}
}

func PromptInput(label string, opts ...GetInputOption) (string, error) {
	input := textinput.New(label)
	for _, opt := range opts {
		input = opt(input)
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

func PromptSelect(label string, choices []string) (int, string, error) {
	sp := selection.New(label, choices)
	sp.PageSize = 4
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
