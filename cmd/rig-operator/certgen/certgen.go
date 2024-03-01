package certgen

import (
	"github.com/spf13/cobra"
)

func CMD() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "tools for generating certificates for webhooks",
	}

	cmd.AddCommand(createCMD())
	patchCMD, err := patchCMD()
	if err != nil {
		return nil, err
	}
	cmd.AddCommand(patchCMD)

	return cmd, nil
}
