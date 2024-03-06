package migrate

import (
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/spf13/cobra"
)

var promptAborted = "prompt aborted"

var (
	skipPlatform bool
	apply        bool
)

var name string

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
	migrate.Flags().StringVar(&name, "name", "", "The name of the capsule to create.\n"+
		"If empty, the capsule name defaults to the service name."+
		" If no service is found, it defaults to the deployment name. \n"+
		"The name will be propagated to the created resources.")
	migrate.Flags().BoolVarP(&apply, "apply", "a", false, "Apply the capsule to the rig platform")
	migrate.MarkFlagsMutuallyExclusive("apply", "skip-platform")

	parent.AddCommand(migrate)
}
