package apichecker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rigdev/rig/cmd/rig-operator/log"
	"github.com/rigdev/rig/pkg/apichecker"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	flagInterval  = "interval"
	flagTimeout   = "timeout"
	flagNamespace = "namespace"
)

func Setup(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "apicheck",
		Short: "checks if rig-operator apis are ready",
		Long:  "checks if rig-operator apis are ready by dry-running a Capsule creation",
		RunE:  apicheck,
	}

	flags := cmd.Flags()
	flags.Duration(flagInterval, time.Second*5, "interval for running check")
	flags.Duration(flagTimeout, time.Minute*2, "timeout for when we should stop checking")
	flags.String(flagNamespace, "rig-system", "namespace where Capsule resource creation will dryrun")

	parent.AddCommand(cmd)
}

func apicheck(cmd *cobra.Command, _ []string) error {
	interval, err := cmd.Flags().GetDuration(flagInterval)
	if err != nil {
		return err
	}
	timeout, err := cmd.Flags().GetDuration(flagTimeout)
	if err != nil {
		return err
	}
	namespace, err := cmd.Flags().GetString(flagNamespace)
	if err != nil {
		return err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
		if err != nil {
			return fmt.Errorf("could not create kubernetes config: %w", err)
		}
	}

	s := scheme.New()
	c, err := client.New(config, client.Options{
		Scheme: s,
	})
	if err != nil {
		return fmt.Errorf("could not create k8s client for apicheck")
	}
	c = client.NewNamespacedClient(client.NewDryRunClient(c), namespace)

	log := log.New(false)

	checker := apichecker.New(c)
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	ticker := time.NewTicker(interval)
	var lastErr error
	for {
		log.Info("checking rig-operator api")
		if err := checker.Check(cmd.Context()); err != nil {
			log.WithValues("error", err).Info("rig-operator api is not ready")
			lastErr = err
		} else {
			log.Info("rig-operator api is ready")
			return nil
		}
		select {
		case <-ctx.Done():
			return lastErr
		case <-ticker.C:
			continue
		}
	}
}
