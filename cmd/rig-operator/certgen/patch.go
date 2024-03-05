package certgen

import (
	"fmt"

	"github.com/rigdev/rig/cmd/rig-operator/log"
	"github.com/spf13/cobra"
)

const (
	flagValidating      = "validating"
	flagMutating        = "mutating"
	flagCRDs            = "crds"
	flagSecretName      = "secret-name"
	flagSecretNamespace = "secret-namespace"
	flagAPIServiceName  = "api-service-name"
	flagWebhookCFGName  = "webhook-cfg-name"
)

func patchCMD() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "patch",
		Args:  cobra.ExactArgs(0),
		Short: "Patch validating/mutating webhook configurations and CRDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			webhookCFGName, err := flags.GetString(flagWebhookCFGName)
			if err != nil {
				return err
			}
			patchValidating, err := flags.GetBool(flagValidating)
			if err != nil {
				return err
			}
			patchMutating, err := flags.GetBool(flagMutating)
			if err != nil {
				return err
			}
			patchCRDs, err := flags.GetBool(flagCRDs)
			if err != nil {
				return err
			}
			secretName, err := flags.GetString(flagSecretName)
			if err != nil {
				return err
			}
			secretNamespace, err := flags.GetString(flagSecretNamespace)
			if err != nil {
				return err
			}

			k8s, err := newK8s()
			if err != nil {
				return err
			}

			log := log.New(false).WithValues(
				"validating", patchValidating,
				"mutating", patchMutating,
				"crds", patchCRDs,
				"secretName", secretName,
				"secretNamespace", secretNamespace,
				"name", webhookCFGName,
			)

			log.Info("getting certificate secret...")
			s, err := k8s.getSecret(cmd.Context(), secretNamespace, secretName)
			if err != nil {
				return fmt.Errorf("could not get secret: %w", err)
			}

			ca := s.Data["ca.crt"]
			if ca == nil {
				return fmt.Errorf("secret %s/%s does not contain ca", secretNamespace, secretName)
			}

			log.Info("found certificate secret")
			if patchValidating && webhookCFGName != "" {
				log.Info("patching validating")
				if err := k8s.patchValidating(cmd.Context(), webhookCFGName, ca); err != nil {
					return err
				}
				log.Info("patched validating")
			}
			if patchMutating && webhookCFGName != "" {
				log.Info("patching mutating")
				if err := k8s.patchMutating(cmd.Context(), webhookCFGName, ca); err != nil {
					return err

				}
				log.Info("patched mutating")
			}
			if patchCRDs {
				log.Info("patching Capsule CRD")
				if err := k8s.patchCRD(cmd.Context(), "capsules.rig.dev", ca); err != nil {
					return err
				}
				log.Info("patched Capsule CRD")
			}

			return nil
		},
	}

	flags := cmd.PersistentFlags()
	flags.Bool(flagValidating, true, "wether to patch ValidatingWebhookConfiguration with given name")
	flags.Bool(flagMutating, true, "wether to patch MutatingWebhookConfiguration with given name")
	flags.Bool(flagCRDs, true, "wether to patch CRDs")
	flags.String(flagWebhookCFGName, "", "Name of *WebhookConfiguration resources")
	flags.String(flagSecretName, "", "Name of certificate secret containing the ca")
	flags.String(flagSecretNamespace, "default", "Namespace of certificate secret containing the ca")

	if err := cobra.MarkFlagRequired(flags, flagSecretName); err != nil {
		return nil, err
	}

	return cmd, nil
}
