package common

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/term"
)

type GroupedFlagSet struct {
	*pflag.FlagSet
	groupName string
	priority  int
}

const groupAnnotation = "rigdev.grouped_flag_set"

func NewGroupedFlagSet(groupName string, priority int) *GroupedFlagSet {
	return &GroupedFlagSet{
		FlagSet:   pflag.NewFlagSet("", pflag.ContinueOnError),
		groupName: groupName,
		priority:  priority,
	}
}

func (g GroupedFlagSet) Finalize() {
	g.VisitAll(func(f *pflag.Flag) {
		if f.Annotations == nil {
			f.Annotations = map[string][]string{}
		}
		f.Annotations[groupAnnotation] = []string{g.groupName, fmt.Sprintf("%v", g.priority)}
	})
}

type flagGroup struct {
	name     string
	flags    *pflag.FlagSet
	priority int
}

func splitFlagsByGroup(flags *pflag.FlagSet) []flagGroup {
	groups := map[string]flagGroup{}
	flags.VisitAll(func(f *pflag.Flag) {
		var group string
		var priority int
		var err error
		if a, ok := f.Annotations[groupAnnotation]; ok {
			group = a[0]
			priority, err = strconv.Atoi(a[1])
			if err != nil {
				panic(err)
			}
		}
		flags, ok := groups[group]
		if !ok {
			flags = flagGroup{
				name:     group,
				flags:    pflag.NewFlagSet("", pflag.ContinueOnError),
				priority: priority,
			}
		}
		flags.flags.AddFlag(f)
		groups[group] = flags
	})

	return slices.SortedFunc(maps.Values(groups), func(g1, g2 flagGroup) int {
		if g1.priority != g2.priority {
			return g2.priority - g1.priority
		}
		return strings.Compare(g1.name, g2.name)
	})
}

func groupedFlagUsages(cmd *pflag.FlagSet) string {
	var buf bytes.Buffer

	groups := splitFlagsByGroup(cmd)
	for idx, group := range groups {
		name := group.name
		if name != "" || len(groups) > 1 {
			if name == "" {
				name = "Other"
			}
			s := fmt.Sprintf("  %s Flags\n", name)
			buf.Write([]byte(s))
		}
		buf.Write([]byte(wrappedFlagUsages(group.flags)))
		if idx != len(groups)-1 {
			buf.Write([]byte("\n"))
		}
	}

	return buf.String()
}

// Stolen from https://github.com/vmware-tanzu/community-edition/blob/138acbf49d492815d7f72055db0186c43888ae15/cli/cmd/plugin/unmanaged-cluster/cmd/utils.go#L74-L113
// Uses the users terminal size or width of 80 if cannot determine users width
//
//nolint:lll
func wrappedFlagUsages(cmd *pflag.FlagSet) string {
	fd := int(os.Stdout.Fd())
	width := 80

	// Get the terminal width and dynamically set
	termWidth, _, err := term.GetSize(fd)
	if err == nil {
		width = termWidth
	}

	return cmd.FlagUsagesWrapped(width - 1)
}
