package migrate

import (
	"fmt"
)

type Warning struct {
	Kind       string
	Name       string
	Field      string
	Warning    string
	Suggestion string
}

func (w *Warning) String() string {
	str := fmt.Sprintf("%s/%s", w.Kind, w.Name)

	if w.Field != "" {
		str += fmt.Sprintf(".%s", w.Field)
	}

	str += fmt.Sprintf(":\nWarning: %s", w.Warning)

	if w.Suggestion != "" {
		str += fmt.Sprintf("\nSugggestion: %s", w.Suggestion)
	}

	return str + "\n-----------"
}
