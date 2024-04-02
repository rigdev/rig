package certgen

import (
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) error {
	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "tools for generating certificates for webhooks",
	}

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Generate ca, server cert and key and store it in a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  create,
	}
	flags := createCmd.PersistentFlags()
	flags.StringP(flagNamespace, "n", "default", "Namespace for certificate secret")
	flags.StringSlice(flagHosts, nil, "IPs and DNS names to include in the certificate")
	cmd.AddCommand(createCmd)

	patchCmd := &cobra.Command{
		Use:   "patch",
		Args:  cobra.ExactArgs(0),
		Short: "Patch validating/mutating webhook configurations and CRDs",
		RunE:  patch,
	}
	flags = patchCmd.PersistentFlags()
	flags.Bool(flagValidating, true, "wether to patch ValidatingWebhookConfiguration with given name")
	flags.Bool(flagMutating, true, "wether to patch MutatingWebhookConfiguration with given name")
	flags.Bool(flagCRDs, true, "wether to patch CRDs")
	flags.String(flagWebhookCFGName, "", "Name of *WebhookConfiguration resources")
	flags.String(flagSecretName, "", "Name of certificate secret containing the ca")
	flags.String(flagSecretNamespace, "default", "Namespace of certificate secret containing the ca")
	if err := cobra.MarkFlagRequired(flags, flagSecretName); err != nil {
		return err
	}
	cmd.AddCommand(patchCmd)
	parent.AddCommand(cmd)

	return nil
}
