package storage

import (
	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c Cmd) createBucket(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	var pid string
	var err error
	if len(args) == 1 {
		id, err := uuid.Parse(args[0])
		if err != nil {
			res, err := c.Rig.Storage().LookupProvider(ctx, &connect.Request[storage.LookupProviderRequest]{
				Msg: &storage.LookupProviderRequest{
					Name: args[0],
				},
			})
			if err != nil {
				return err
			}
			pid = res.Msg.GetProviderId()
		} else {
			pid = id.String()
		}
	} else {
		res, err := c.Rig.Storage().ListProviders(ctx, &connect.Request[storage.ListProvidersRequest]{
			Msg: &storage.ListProvidersRequest{},
		})
		if err != nil {
			return err
		}

		if len(res.Msg.GetProviders()) == 0 {
			return errors.NotFoundErrorf("no providers found")
		}

		if len(res.Msg.GetProviders()) == 1 {
			pid = res.Msg.GetProviders()[0].GetProviderId()
		} else {
			// Ask the user to choose a provider.
			providerNames := make([]string, len(res.Msg.GetProviders()))
			for i, p := range res.Msg.GetProviders() {
				providerNames[i] = p.GetName()
			}
			i, _, err := common.PromptSelect("Select provider:", providerNames)
			if err != nil {
				return err
			}

			pid = res.Msg.GetProviders()[i].GetProviderId()
		}
	}

	if name == "" {
		name, err = common.PromptInput("Bucket name:", ValidateBucketNameOpt)
		if err != nil {
			return err
		}
	} else if err := ValidateBucketName(name); err != nil {
		return err
	}

	if providerBucketName == "" {
		providerBucketName, err = common.PromptInput("Provider bucket name:", ValidateBucketNameOpt, common.InputDefaultOpt(name))
		if err != nil {
			return err
		}
	} else if err := ValidateBucketName(providerBucketName); err != nil {
		return err
	}

	_, err = c.Rig.Storage().CreateBucket(ctx, &connect.Request[storage.CreateBucketRequest]{
		Msg: &storage.CreateBucketRequest{
			Bucket:         name,
			ProviderBucket: providerBucketName,
			Region:         region,
			ProviderId:     pid,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Bucket created: ", name)

	return nil
}
