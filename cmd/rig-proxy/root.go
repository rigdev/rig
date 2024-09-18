package main

import (
	"os"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/rig-proxy/tunnel"
	"github.com/rigdev/rig/pkg/build"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/yaml"
)

var verbose = false

func createRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "rig-proxy",
		RunE: func(_ *cobra.Command, _ []string) error {
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(os.Stderr),
				zapcore.DebugLevel,
			)
			logger := zap.New(core)

			p := tunnel.New(logger)

			bs, err := os.ReadFile("/capsule.yaml")
			if err != nil {
				logger.Error("error reading proxy config", zap.Error(err), zap.String("path", "/capsule.yaml"))
				return err
			}

			var cfg platformv1.HostCapsule
			if err := yaml.Unmarshal(bs, &cfg); err != nil {
				logger.Error("error reading proxy config", zap.Error(err), zap.String("path", "/capsule.yaml"))
				return err
			}

			for _, hostIf := range cfg.GetNetwork().GetHostInterfaces() {
				if err := p.AddHostInterface(hostIf); err != nil {
					return err
				}
			}

			for _, capIf := range cfg.GetNetwork().GetCapsuleInterfaces() {
				if err := p.AddCapsuleInterface(capIf); err != nil {
					return err
				}
			}

			return p.Serve(cfg.GetNetwork().GetTunnelPort())
		},
	}

	pflags := cmd.PersistentFlags()
	pflags.BoolVarP(&verbose, "verbose", "v", false, "enable verbose error logging")

	cmd.AddCommand(build.VersionCommand())

	return cmd
}
