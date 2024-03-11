package migrate

import (
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/spf13/cobra"
)

const (
	promptAborted = "prompt aborted"
)

var (
	skipPlatform bool
	apply        bool
)

var nameOrigin CapsuleName

func Setup(parent *cobra.Command) {
	migrate := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate you kubernetes deployments to Rig Capsules",
		RunE:  base.Register(migrate),
	}

	migrate.Flags().BoolVar(&skipPlatform,
		"skip-platform",
		false,
		`Do not perform a platform dryrun to compare to k8s resources from rig-platform. Instead just compare the
resources created from a capsulespec
		If false:
			- The rig platform must be running
			- A valid rig-cli config and context must be set either according to defaults or through flags
			- Valid access- and refresh tokens must be provided in the context. Otherwise a login is prompted.`,
	)
	migrate.PersistentFlags().StringVarP(&base.Flags.Namespace, "namespace", "n", "", "The k8s namespace to migrate from")
	migrate.PersistentFlags().StringVarP(&base.Flags.Project, "project", "p", "", "The project to migrate to")
	migrate.Flags().Var(&nameOrigin, "name-origin", "From where to inherit the name of the capsule. One of `service,deployment,input`."+
		" Default is service, if one exists, otherwise deployment.")
	migrate.Flags().BoolVarP(&apply, "apply", "a", false, "Apply the capsule to the rig platform")
	migrate.MarkFlagsMutuallyExclusive("apply", "skip-platform")

	parent.AddCommand(migrate)
}
