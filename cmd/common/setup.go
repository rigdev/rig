package common

import (
	"github.com/spf13/cobra"
)

// Identical to the default cobra usage template,
// but utilizes groupedFlagsUsage to group flags
var usageTemplate = `Usage:
{{- if .Runnable}}
  {{.UseLine}}
{{- end}}

{{- if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]
{{- end}}

{{- if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}
{{- end}}

{{- if .HasExample}}

Examples:
{{.Example}}
{{- end}}

{{- if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}
{{- end}}

{{- if .HasAvailableLocalFlags}}

Flags:
{{groupedFlagUsages .LocalFlags | trimTrailingWhitespaces}}
{{- end}}

{{- if .HasAvailableInheritedFlags}}

Global Flags:
{{groupedFlagUsages .InheritedFlags | trimTrailingWhitespaces}}
{{- end}}

{{- if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func SetupRoot(cmd *cobra.Command) {
	cobra.AddTemplateFunc("groupedFlagUsages", groupedFlagUsages)
	cmd.SetUsageTemplate(usageTemplate)

}
