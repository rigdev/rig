package certgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type k8s struct {
	client           kubernetes.Interface
	extensionsClient apiext.Interface
}

func newK8s() (*k8s, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
		if err != nil {
			return nil, fmt.Errorf("could not create kubernetes config: %w", err)
		}
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes client: %w", err)
	}

	extcs, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes apiextensions client: %w", err)
	}

	return &k8s{
		client:           cs,
		extensionsClient: extcs,
	}, nil
}

func (k *k8s) getSecret(ctx context.Context, namespace, name string) (*v1.Secret, error) {
	s, err := k.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, err
		}
		return nil, fmt.Errorf("could not get certificate secret %s/%s: %w", namespace, name, err)
	}

	return s, nil
}

func (k *k8s) createSecret(ctx context.Context, namespace, name string, certs *Certs) error {
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Type: v1.SecretTypeTLS,
		Data: map[string][]byte{
			"ca.crt":  certs.CA,
			"tls.crt": certs.Cert,
			"tls.key": certs.Key,
		},
	}

	if _, err := k.client.CoreV1().Secrets(namespace).Create(ctx, s, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("could not create secret: %w", err)
	}
	return nil
}

func (k *k8s) patchValidating(ctx context.Context, name string, ca []byte) error {
	vwc, err := k.client.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get ValidatingWebhookConfiguration %s: %w", name, err)
	}

	for i := range vwc.Webhooks {
		vwc.Webhooks[i].ClientConfig.CABundle = ca
	}

	if _, err := k.client.
		AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Update(ctx, vwc, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("could not update ValidatingWebhookConfiguration %s: %w", name, err)
	}
	return nil
}

func (k *k8s) patchMutating(ctx context.Context, name string, ca []byte) error {
	mwc, err := k.client.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get MutatingWebhookConfiguration %s: %w", name, err)
	}

	for i := range mwc.Webhooks {
		mwc.Webhooks[i].ClientConfig.CABundle = ca
	}

	if _, err := k.client.
		AdmissionregistrationV1().
		MutatingWebhookConfigurations().
		Update(ctx, mwc, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("could not update MutatingWebhookConfiguration %s: %w", name, err)
	}
	return nil
}

func (k *k8s) patchCRD(ctx context.Context, name string, ca []byte) error {
	crd, err := k.extensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get CustomResourceDefinition %s: %w", name, err)
	}

	if crd.Spec.Conversion != nil &&
		crd.Spec.Conversion.Webhook != nil &&
		crd.Spec.Conversion.Webhook.ClientConfig != nil {
		crd.Spec.Conversion.Webhook.ClientConfig.CABundle = ca
	}

	if _, err := k.extensionsClient.
		ApiextensionsV1().
		CustomResourceDefinitions().
		Update(ctx, crd, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("could not update CRD %s: %w", name, err)
	}

	return nil
}
