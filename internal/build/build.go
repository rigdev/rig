package build

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func Date() string {
	return date
}

func VersionString() string {
	return fmt.Sprintf("rig %s", version)
}

func VersionStringFull() string {
	return fmt.Sprintf("%s\ncommit: %s\ndate: %s", VersionString(), commit, date)
}

func VersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			full, err := cmd.Flags().GetBool("full")
			if err != nil {
				return err
			}

			if full {
				fmt.Println(VersionStringFull())
			} else {
				fmt.Println(VersionString())
			}

			return nil
		},
	}

	cmd.Flags().BoolP("full", "v", false, "print full version")

	return cmd
}
