package migrate

import (
	"fmt"
)

type Warning struct {
	Kind       string
	Name       string
	Warning    string
	Suggestion string
}

func (w *Warning) String() string {
	str := fmt.Sprintf(
		"%s/%s: %s", w.Kind, w.Name, w.Warning)
	if w.Suggestion != "" {
		str += fmt.Sprintf(
			"\nSugggestion: %s" +
				w.Suggestion)
	}
	return str + "\n------------------------------"
}
