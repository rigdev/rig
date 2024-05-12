package scheme

import (
	certv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	configv1alpha1 "github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	batchv1 "k8s.io/api/batch/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// New returns a new *runtime.Scheme configured with the types we use in
// this project.
func New() *runtime.Scheme {
	s := runtime.NewScheme()

	utilruntime.Must(clientsetscheme.AddToScheme(s))
	utilruntime.Must(certv1.AddToScheme(s))
	utilruntime.Must(monitorv1.AddToScheme(s))
	utilruntime.Must(batchv1.AddToScheme(s))

	utilruntime.Must(configv1alpha1.AddToScheme(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(v1alpha2.AddToScheme(s))
	utilruntime.Must(apiextensionsv1.AddToScheme(s))
	utilruntime.Must(vpav1.AddToScheme(s))

	utilruntime.Must(metricsv1beta1.AddToScheme(s))

	return s
}
