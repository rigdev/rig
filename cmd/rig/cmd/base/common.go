package base

import "github.com/spf13/cobra"

const (
	RFC3339NanoFixed  = "2006-01-02T15:04:05.000000000Z07:00"
	RFC3339MilliFixed = "2006-01-02T15:04:05.000Z07:00"
)

// ExecutePersistentPreRunERecursively executes any registrered PersistentPreRunE functions
// on the parent-chain from the given cmd.
// The 'persistent' in PersistentPreRunE is only if no child-commands also sets the function.
func ExecutePersistentPreRunERecursively(cmd *cobra.Command, args []string) error {
	if !cmd.HasParent() {
		return nil
	}
	cmd = cmd.Parent()

	if err := ExecutePersistentPreRunERecursively(cmd, args); err != nil {
		return err
	}

	if cmd.PersistentPreRunE != nil {
		if err := cmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}
	}

	return nil
}
