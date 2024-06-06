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
	"github.com/go-logr/logr"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline/pipelineconnect"
	"github.com/rigdev/rig/cmd/rig-operator/apichecker"
	"github.com/rigdev/rig/cmd/rig-operator/certgen"
	"github.com/rigdev/rig/cmd/rig-operator/log"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/build"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/handler/api/capabilities"
	"github.com/rigdev/rig/pkg/handler/api/pipeline"
	"github.com/rigdev/rig/pkg/manager"
	"github.com/rigdev/rig/pkg/scheme"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"github.com/rigdev/rig/pkg/service/objectstatus"
	svcpipeline "github.com/rigdev/rig/pkg/service/pipeline"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
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
	app := fx.New(
		fx.Provide(
			ctrl.GetConfigOrDie,
			scheme.New,
			func(cfg config.Service) logr.Logger {
				log := log.New(cfg.Operator().DevModeEnabled)
				ctrl.SetLogger(log)
				return log
			},
			func(scheme *runtime.Scheme) (config.Service, *v1alpha1.OperatorConfig, error) {
				cfgFile, err := cmd.Flags().GetString(flagConfigFile)
				if err != nil {
					return nil, nil, err
				}

				cfg, err := config.NewService(scheme, cfgFile)
				if err != nil {
					return nil, nil, err
				}

				return cfg, cfg.Operator(), nil
			},
			func(lc fx.Lifecycle, log logr.Logger) context.Context {
				ctx, cancel := context.WithCancel(cmd.Context())
				lc.Append(fx.StopHook(cancel))

				go func() {
					signals := make(chan os.Signal, 1)
					signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

					<-signals

					log.Info("stopping manager...")

					// Stop everything in progress, doing a graceful shutdown.
					cancel()
				}()

				return ctx
			},
			func(restConfig *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
				return client.New(restConfig, client.Options{
					Scheme: scheme,
				})
			},
			func(restConfig *rest.Config) (clientset.Interface, error) {
				return clientset.NewForConfig(restConfig)
			},
			func(cc clientset.Interface) discovery.DiscoveryInterface {
				return cc.Discovery()
			},
			plugin.NewManager,
			svccapabilities.NewService,
			capabilities.NewHandler,
			svcpipeline.NewService,
			objectstatus.NewService,
			pipeline.NewHandler,
			manager.New,
		),
		fx.Invoke(
			func(
				log logr.Logger,
				ctx context.Context,
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
					MaxHeaderBytes:    8 * 1024, // 8KiB
				}

				lc.Append(fx.StopHook(func(ctx context.Context) {
					ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
					defer cancel()
					_ = srv.Shutdown(ctx)
				}))

				go func() {
					log.Info("starting GRPC server")
					if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
						log.Error(err, "could not start GRPC server")
						_ = sh.Shutdown(fx.ExitCode(1))
					}
				}()

				ctx, cancel := context.WithCancel(ctx)
				done := make(chan struct{})
				lc.Append(fx.StopHook(func() {
					cancel()
					<-done
				}))

				go func() {
					defer close(done)
					log.Info("starting manager server")
					if err := mgr.Start(ctx); err != nil {
						log.Error(err, "failed starting manager")
						_ = sh.Shutdown(fx.ExitCode(1))
						return
					}

					_ = sh.Shutdown(fx.ExitCode(0))

					log.Info("manager stopped")
				}()
			},
		),
	)

	app.Run()

	return app.Err()
}
