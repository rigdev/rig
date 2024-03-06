package migrate

import (
	"fmt"
)

type Warning struct {
	Kind    string
	Name    string
	Warning string
}

func (w *Warning) String() string {
	return fmt.Sprintf(
		"%s/%s: %s"+
			"\n----------------------------------------",
		w.Kind, w.Name, w.Warning)
}
