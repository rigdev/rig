package base

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline/pipelineconnect"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorClient struct {
	Pipeline     pipelineconnect.ServiceClient
	Capabilities capabilitiesconnect.ServiceClient
}

func NewOperatorClient(ctx context.Context, cc client.Client, cfg *rest.Config) (*OperatorClient, error) {
	pods := &v1.PodList{}
	if err := cc.List(ctx, pods, client.InNamespace("rig-system"), client.MatchingLabels{
		"app.kubernetes.io/name": "rig-operator",
	}); err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no `rig-operator` pods found")
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	url := cs.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pods.Items[0].GetNamespace()).
		Name(pods.Items[0].GetName()).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create round tripper")
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	rdy := make(chan struct{}, 1)
	errs := make(chan error, 1)
	pw, err := portforward.New(dialer, []string{"0:9000"}, nil, rdy, io.Discard, io.Discard)
	if err != nil {
		return nil, err
	}

	go func() {
		errs <- pw.ForwardPorts()
	}()

	select {
	case <-rdy:
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errs:
		return nil, err
	}

	ps, err := pw.GetPorts()
	if err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("http://localhost:%d/", ps[0].Local)

	return &OperatorClient{
		Pipeline:     pipelineconnect.NewServiceClient(http.DefaultClient, baseURL),
		Capabilities: capabilitiesconnect.NewServiceClient(http.DefaultClient, baseURL),
	}, nil
}

func GetOperatorConfig(
	ctx context.Context,
	operatorClient *OperatorClient,
	scheme *runtime.Scheme,
) (*v1alpha1.OperatorConfig, error) {
	var cfgYAML string
	if Flags.OperatorConfig == "" {
		cfgResp, err := operatorClient.Capabilities.GetConfig(ctx, connect.NewRequest(&capabilities.GetConfigRequest{}))
		if err != nil {
			return nil, err
		}
		cfgYAML = cfgResp.Msg.GetYaml()
	} else {
		bytes, err := os.ReadFile(Flags.OperatorConfig)
		if err != nil {
			return nil, err
		}
		cfgYAML = string(bytes)
	}
	return obj.DecodeIntoT([]byte(cfgYAML), &v1alpha1.OperatorConfig{}, scheme)
}
