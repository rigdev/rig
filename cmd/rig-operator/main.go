package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig/pkg/build"
	"github.com/rigdev/rig/pkg/handler/api/capabilities"
	"github.com/rigdev/rig/pkg/manager"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	flagConfigFile = "config-file"
)

func main() {
	c := &cobra.Command{
		Use:   "rig-operator",
		Short: "operator for the rig.dev CRDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args)
		},
	}

	flags := c.PersistentFlags()
	flags.StringP(flagConfigFile, "c", "/etc/rig-operator/config.yaml", "path to rig-operator config file")

	c.AddCommand(build.VersionCommand())

	ctx := context.Background()
	if err := c.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	cfgFile, err := cmd.Flags().GetString(flagConfigFile)
	if err != nil {
		return err
	}

	scheme := manager.NewScheme()

	cfg, err := config.NewService(cfgFile, scheme)
	if err != nil {
		return err
	}

	log := zap.New(zap.UseDevMode(cfg.Get().DevModeEnabled))

	ctrl.SetLogger(log)

	mgr, err := manager.NewManager(cfg, scheme)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		<-signals

		// Stop everything in progress, doing a graceful shutdown.
		cancel()
	}()

	capabilitiesSvc := svccapabilities.NewService(cfg)
	capabilitiesH := capabilities.NewHandler(capabilitiesSvc)

	mux := http.NewServeMux()
	mux.Handle(capabilitiesconnect.NewServiceHandler(
		capabilitiesH,
	))
	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(capabilitiesconnect.ServiceName),
	))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(capabilitiesconnect.ServiceName),
	))

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
		Addr:              ":9000",
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		MaxHeaderBytes:    8 * 1024, // 8KiB
	}

	go func() {
		log.Info("starting GRPC server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(err, "could not start GRPC server")
			os.Exit(1)
		}
	}()

	log.Info("starting manager server")
	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "failed starting manager")
		return err
	}

	log.Info("manager stopped")

	log.Info("stopping GRPC server...")
	grpcCTX, grpcCancel := context.WithTimeout(cmd.Context(), time.Second)
	defer grpcCancel()
	if err := srv.Shutdown(grpcCTX); err != nil {
		return err
	}

	log.Info("stopping manager...")
	return nil
}
