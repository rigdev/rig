package migrate

import (
	"github.com/rigdev/rig/cmd/common"
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
	apply             bool
	deploymentName    string
	annotations       map[string]string
	keepEnvConfigMaps bool
)

var nameOrigin CapsuleName

type Cmd struct {
	fx.In

	OperatorClient *base.OperatorClient
	K8s            client.Client
	K8sReader      client.Reader
	Scheme         *runtime.Scheme
	Prompter       common.Prompter
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

	migrate.Flags().StringVarP(&base.Flags.Environment, "environment", "e", "", "The environment to migrate to")
	migrate.Flags().StringVarP(&base.Flags.Project, "project", "p", "", "The project to migrate to")
	migrate.Flags().Var(&nameOrigin, "name-origin",
		"From where to inherit the name of the capsule. One of `service,deployment,input`."+
			" Default is service, if one exists, otherwise deployment.")
	migrate.Flags().BoolVarP(&apply, "apply", "a", false, "Apply the capsule to the rig platform")
	migrate.Flags().StringVar(&deploymentName, "deployment", "",
		"The deployment to migrate. If not set, a list of deployments will be prompted for.")
	migrate.Flags().StringToStringVarP(
		&annotations,
		"annotation", "A", nil,
		"annotations to add to the Capsule of the format `key=value`."+
			" Can for example be used to target migration with a specific plugin",
	)
	migrate.Flags().BoolVar(&keepEnvConfigMaps, "keep-env-config-maps", false,
		"Keep envFrom to ConfigMaps after migration. "+
			"By default, ConfigMaps are read and added to the Capsule. Changing this behavior is useful "+
			"if an external tool is generating the ConfigMaps.",
	)

	parent.AddCommand(migrate)
}
