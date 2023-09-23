package storage

import (
	"context"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	storageRecursive bool
	linkBuckets      bool
	outputJson       bool

	GCS   bool
	S3    bool
	Minio bool
)

var (
	name               string
	credsFilePath      string
	accessKey          string
	secretKey          string
	region             string
	endpoint           string
	providerBucketName string
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
}

func (c Cmd) Setup(parent *cobra.Command) {
	storage := &cobra.Command{
		Use: "storage",
	}

	cp := &cobra.Command{
		Use:     "copy from to",
		Aliases: []string{"cp"},
		Short:   "Copy files to and from buckets",
		Args:    cobra.ExactArgs(2),
		RunE:    c.cp,
	}
	cp.PersistentFlags().BoolVarP(&storageRecursive, "recursive", "r", false, "if copy should be recursive")
	storage.AddCommand(cp)

	ls := &cobra.Command{
		Use:     "list [path]",
		Aliases: []string{"ls"},
		Short:   "List buckets and objects",
		Args:    cobra.MaximumNArgs(1),
		RunE:    c.ls,
	}
	ls.PersistentFlags().BoolVarP(&storageRecursive, "recursive", "r", false, "if listing should be recursive. Does only work for listing within a single bucket")
	ls.Flags().BoolVar(&outputJson, "json", false, "output as json")
	storage.AddCommand(ls)

	createBucket := &cobra.Command{
		Use:   "create-bucket [provider-name]",
		Short: "Create a new bucket",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.createBucket,
	}
	createBucket.Flags().StringVarP(&name, "name", "n", "", "name of the bucket")
	createBucket.Flags().StringVarP(&providerBucketName, "provider-bucket-name", "p", "", "name of the bucket on the provider")
	createBucket.Flags().StringVarP(&region, "region", "r", "", "region of the bucket")
	storage.AddCommand(createBucket)

	deleteBucket := &cobra.Command{
		Use:   "delete-bucket [bucket-name]",
		Short: "Delete a bucket",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.deleteBucket,
	}
	storage.AddCommand(deleteBucket)

	unlinkBucket := &cobra.Command{
		Use:   "unlink-bucket [bucket-name]",
		Short: "Unlink a bucket",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.unlinkBucket,
	}
	storage.AddCommand(unlinkBucket)

	getObject := &cobra.Command{
		Use:   "get-object [path]",
		Short: "Get an object",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getObject,
	}
	getObject.Flags().BoolVar(&outputJson, "json", false, "output as json")
	storage.AddCommand(getObject)

	getBucket := &cobra.Command{
		Use:   "get-bucket [bucket]",
		Short: "Get a bucket",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getBucket,
	}
	getBucket.Flags().BoolVar(&outputJson, "json", false, "output as json")
	storage.AddCommand(getBucket)

	deleteObject := &cobra.Command{
		Use:   "delete-object [path]",
		Short: "Delete an object",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.deleteObject,
	}
	storage.AddCommand(deleteObject)

	createProvider := &cobra.Command{
		Use:   "create-provider",
		Short: "Create a new provider",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.createProvider,
	}
	createProvider.Flags().StringVarP(&name, "name", "n", "", "name of the provider")

	createProvider.Flags().BoolVar(&GCS, "gcs", false, "if the provider should be a GCS provider")
	createProvider.Flags().BoolVar(&S3, "s3", false, "if the provider should be a S3 provider")
	createProvider.Flags().BoolVar(&Minio, "minio", false, "if the provider should be a Minio provider")
	createProvider.MarkFlagsMutuallyExclusive("gcs", "s3", "minio")

	createProvider.Flags().StringVarP(&credsFilePath, "creds-file", "c", "", "path to the GCS credentials file")
	createProvider.MarkFlagsRequiredTogether("gcs", "creds-file")

	createProvider.Flags().StringVarP(&accessKey, "access-key", "a", "", "access key for the provider")
	createProvider.Flags().StringVarP(&secretKey, "secret-key", "s", "", "secret key for the provider")
	createProvider.Flags().StringVarP(&region, "region", "r", "", "region for the provider")
	createProvider.Flags().StringVarP(&endpoint, "endpoint", "e", "", "endpoint for the provider")

	createProvider.MarkFlagsRequiredTogether("s3", "region")
	createProvider.MarkFlagsRequiredTogether("minio", "endpoint")
	createProvider.MarkFlagsRequiredTogether("access-key", "secret-key")

	createProvider.Flags().BoolVarP(&linkBuckets, "link-buckets", "l", false, "if buckets should be linked to the provider")
	storage.AddCommand(createProvider)

	listProviders := &cobra.Command{
		Use:   "list-providers",
		Short: "List all providers",
		Args:  cobra.NoArgs,
		RunE:  c.listProviders,
	}
	listProviders.Flags().IntVarP(&limit, "limit", "l", 10, "limit the number of groups to return")
	listProviders.Flags().IntVarP(&offset, "offset", "o", 0, "offset the number of groups to return")
	listProviders.Flags().BoolVar(&outputJson, "json", false, "output as json")
	storage.AddCommand(listProviders)

	parent.AddCommand(storage)

	GetProvider := &cobra.Command{
		Use:   "get-provider [id | name]",
		Short: "Get a provider",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getProvider,
	}
	GetProvider.Flags().BoolVar(&outputJson, "json", false, "output as json")
	storage.AddCommand(GetProvider)

	DeleteProvider := &cobra.Command{
		Use:   "delete-provider [id | name]",
		Short: "Delete a provider",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.deleteProvider,
	}
	storage.AddCommand(DeleteProvider)
}

var ValidateBucketName = func(input string) error {
	// check if length of input is greater than 3 and less than 63
	if len(input) < 3 || len(input) > 63 {
		return errors.InvalidArgumentErrorf("bucket name must be between 3 and 63 characters long")
	}
	// check if input contains only lowercase letters, numbers, dashes and underscores
	for _, c := range input {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return errors.InvalidArgumentErrorf("bucket name must only contain lowercase letters, numbers, dashes and underscores")
		}
	}
	// input must start and end with a lowercase letter or number
	if !(input[0] >= 'a' && input[0] <= 'z') && !(input[0] >= '0' && input[0] <= '9') {
		return errors.InvalidArgumentErrorf("bucket name must start with a lowercase letter or number")
	}

	if !(input[len(input)-1] >= 'a' && input[len(input)-1] <= 'z') && !(input[len(input)-1] >= '0' && input[len(input)-1] <= '9') {
		return errors.InvalidArgumentErrorf("bucket name must end with a lowercase letter or number")
	}
	return nil
}

var ValidateBucketNameOpt = func(inp *textinput.TextInput) {
	inp.Validate = ValidateBucketName
}
