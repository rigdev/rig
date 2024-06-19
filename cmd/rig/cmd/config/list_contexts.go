package config

import (
	"github.com/fatih/color"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *CmdNoScope) listContexts(_ *cobra.Command, _ []string) error {
	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(c.Cfg.Contexts, flags.Flags.OutputType)
	}

	type context struct {
		current     bool
		name        string
		host        string
		loggedIn    bool
		project     string
		environment string
	}
	var contexts []context
	cfg := c.Cfg
	for _, c := range cfg.Contexts {
		context := context{
			current:     c.Name == cfg.CurrentContextName,
			name:        c.Name,
			project:     c.ProjectID,
			environment: c.EnvironmentID,
		}

		if service, err := cfg.GetService(c.Name); err == nil {
			context.host = service.Server
		}
		if user, err := cfg.GetUser(c.Name); err == nil {
			if a := user.Auth; a != nil {
				context.loggedIn = (a.UserID != "" && !uuid.UUID(a.UserID).IsNil()) && a.AccessToken != "" && a.RefreshToken != ""
			}
		}
		contexts = append(contexts, context)
	}
	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("Current", "Name", "Host", "Logged In", "Project", "Environment")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, c := range contexts {
		current := ""
		if c.current {
			current = "✔"
		}
		loggedIn := "✗"
		if c.loggedIn {
			loggedIn = "✔"
		}
		tbl.AddRow(current, c.name,
			common.StringOr(c.host, "-"), loggedIn, common.StringOr(c.project, "-"),
			common.StringOr(c.environment, "-"),
		)
	}
	tbl.Print()

	return nil
}
