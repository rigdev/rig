package common

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rigdev/rig/pkg/utils"
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
		s := flagUsages(group.flags)
		buf.Write([]byte(s))
		if idx != len(groups)-1 {
			buf.Write([]byte("\n"))
		}
	}

	return buf.String()
}

const maxWidth = 120

func flagUsages(flags *pflag.FlagSet) string {
	fd := int(os.Stdout.Fd())
	width := 80
	termWidth, _, err := term.GetSize(fd)
	if err == nil {
		width = min(termWidth, maxWidth)
	}

	var buffer bytes.Buffer
	indent := "  "
	numFlags := 0
	flags.VisitAll(func(_ *pflag.Flag) { numFlags++ })
	idx := 0
	flags.VisitAll(func(flag *pflag.Flag) {
		idx++
		if flag.Hidden {
			return
		}
		header := ""
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			header = fmt.Sprintf("%s-%s, --%s", indent, flag.Shorthand, flag.Name)
		} else {
			header = fmt.Sprintf("%s--%s", indent, flag.Name)
		}
		if flag.NoOptDefVal != "" {
			switch flag.Value.Type() {
			case "string":
				header += fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal)
			case "bool":
				if flag.NoOptDefVal != "true" {
					header += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			case "count":
				if flag.NoOptDefVal != "+1" {
					header += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			default:
				header += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
			}
		}
		color.New(color.Bold).Fprintln(&buffer, header+":")
		buffer.WriteString(utils.WordWrap(flag.Usage, width, indent+indent))
		if idx < numFlags {
			buffer.WriteString("\n")
		}
	})
	return buffer.String()
}
