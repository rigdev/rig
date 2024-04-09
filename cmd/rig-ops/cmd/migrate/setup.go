package migrate

import (
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	promptAborted = "prompt aborted"
)

var (
	skipPlatform   bool
	apply          bool
	deploymentName string
	annotations    map[string]string
)

var nameOrigin CapsuleName

type Cmd struct {
	fx.In

	OperatorClient *base.OperatorClient
	K8s            client.Client
	K8sReader      client.Reader
	Scheme         *runtime.Scheme
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	migrate := &cobra.Command{
		Use:               "migrate",
		Short:             "Migrate you kubernetes deployments to Rig Capsules",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.migrate),
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
	migrate.Flags().StringVarP(&base.Flags.Namespace, "namespace", "n", "", "The k8s namespace to migrate from")
	migrate.Flags().StringVarP(&base.Flags.Project, "project", "p", "", "The project to migrate to")
	migrate.Flags().Var(&nameOrigin, "name-origin",
		"From where to inherit the name of the capsule. One of `service,deployment,input`."+
			" Default is service, if one exists, otherwise deployment.")
	migrate.Flags().BoolVarP(&apply, "apply", "a", false, "Apply the capsule to the rig platform")
	migrate.MarkFlagsMutuallyExclusive("apply", "skip-platform")
	migrate.Flags().StringVar(&deploymentName, "deployment", "",
		"The deployment to migrate. If not set, a list of deployments will be prompted for.")
	migrate.Flags().StringToStringVarP(
		&annotations,
		"annotation", "A", nil,
		"annotations to add to the Capsule of the format `key=value`."+
			" Can for example be used to target migration with a specific plugin",
	)

	parent.AddCommand(migrate)
}
