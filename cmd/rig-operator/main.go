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

	"connectrpc.com/grpcreflect"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline/pipelineconnect"
	"github.com/rigdev/rig/cmd/rig-operator/apichecker"
	"github.com/rigdev/rig/cmd/rig-operator/certgen"
	"github.com/rigdev/rig/cmd/rig-operator/log"
	"github.com/rigdev/rig/pkg/build"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/handler/api/capabilities"
	"github.com/rigdev/rig/pkg/handler/api/pipeline"
	"github.com/rigdev/rig/pkg/manager"
	"github.com/rigdev/rig/pkg/scheme"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	svcpipeline "github.com/rigdev/rig/pkg/service/pipeline"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8smanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	flagConfigFile = "config-file"
)

func main() {
	cmd := &cobra.Command{
		Use:   "rig-operator",
		Short: "operator for the rig.dev CRDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	flags := cmd.PersistentFlags()
	flags.StringP(flagConfigFile, "c", "/etc/rig-operator/config.yaml", "path to rig-operator config file")

	cmd.AddCommand(build.VersionCommand())
	if err := certgen.Setup(cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	apichecker.Setup(cmd)
	pluginSetup(cmd)

	ctx := context.Background()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, _ []string) error {
	cfgFile, err := cmd.Flags().GetString(flagConfigFile)
	if err != nil {
		return err
	}

	scheme := scheme.New()

	cfg, err := config.NewService(scheme, cfgFile)
	if err != nil {
		return err
	}

	log := log.New(cfg.Operator().DevModeEnabled)
	ctrl.SetLogger(log)

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		<-signals

		// Stop everything in progress, doing a graceful shutdown.
		cancel()
	}()

	restConfig := ctrl.GetConfigOrDie()

	app := fx.New(
		fx.Supply(scheme, log),
		fx.Provide(
			func() config.Service {
				return cfg
			},
			func() context.Context {
				return ctx
			},
			func(scheme *runtime.Scheme) (client.Client, error) {
				return client.New(restConfig, client.Options{
					Scheme: scheme,
				})
			},
			func() (clientset.Interface, error) {
				return clientset.NewForConfig(restConfig)
			},
			func(cc clientset.Interface) discovery.DiscoveryInterface {
				return cc.Discovery()
			},
			plugin.NewManager,
			svccapabilities.NewService,
			capabilities.NewHandler,
			svcpipeline.NewService,
			pipeline.NewHandler,
			manager.New,
		),
		fx.Invoke(func(
			mgr k8smanager.Manager,
			lc fx.Lifecycle,
			sh fx.Shutdowner,
			cap capabilitiesconnect.ServiceHandler,
			pip pipelineconnect.ServiceHandler,
		) {
			mux := http.NewServeMux()
			mux.Handle(capabilitiesconnect.NewServiceHandler(cap))
			mux.Handle(pipelineconnect.NewServiceHandler(pip))
			mux.Handle(grpcreflect.NewHandlerV1(grpcreflect.NewStaticReflector(
				capabilitiesconnect.ServiceName,
				pipelineconnect.ServiceName,
			)))
			mux.Handle(grpcreflect.NewHandlerV1Alpha(grpcreflect.NewStaticReflector(
				capabilitiesconnect.ServiceName,
				pipelineconnect.ServiceName,
			)))

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

			lc.Append(fx.StopHook(srv.Shutdown))

			go func() {
				log.Info("starting GRPC server")
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error(err, "could not start GRPC server")
					sh.Shutdown(fx.ExitCode(1))
				}
			}()

			go func() {
				log.Info("starting manager server")
				if err := mgr.Start(ctx); err != nil {
					log.Error(err, "failed starting manager")
					sh.Shutdown(fx.ExitCode(1))
					return
				}

				sh.Shutdown(fx.ExitCode(1))

				log.Info("manager stopped")
			}()
		}),
	)

	app.Run()

	log.Info("stopping manager...")
	return app.Err()
}
