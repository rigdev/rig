package config

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) viewConfig(cmd *cobra.Command, _ []string) error {
	var outputType common.OutputType
	if flags.Flags.OutputType == common.OutputTypePretty {
		outputType = common.OutputTypeYAML
	} else {
		outputType = flags.Flags.OutputType
	}

	var toPrint string
	var err error
	if minify {
		toPrint, err = common.Format(c.Scope.GetCfg().Minify(), outputType)
		if err != nil {
			return err
		}
	} else {
		toPrint, err = common.Format(c.Scope.GetCfg(), outputType)
		if err != nil {
			return err
		}
	}

	cmd.Println(toPrint)
	return nil
}
