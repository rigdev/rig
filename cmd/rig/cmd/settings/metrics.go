package settings

import (
	"context"
	"fmt"
	"os"

	"connectrpc.com/connect"
	rollout_api "github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func (c *Cmd) updateMetrics(ctx context.Context, _ *cobra.Command, args []string) error {

	var updates []*settings_api.Update
	for _, a := range addMetric {
		data, err := os.ReadFile(a)
		if err != nil {
			return err
		}

		data, err = yaml.YAMLToJSON(data)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		m := &rollout_api.Metric{}
		if err := protojson.Unmarshal(data, m); err != nil {
			return fmt.Errorf("failed to decode metric from %s: %w", a, err)
		}
		updates = append(updates, &settings_api.Update{
			Field: &settings_api.Update_AddRolloutMetric{
				AddRolloutMetric: m,
			},
		})
	}
	for _, r := range removeMetric {
		updates = append(updates, &settings_api.Update{
			Field: &settings_api.Update_RemoveRolloutMetric_{
				RemoveRolloutMetric: &settings_api.Update_RemoveRolloutMetric{
					Name: r,
				},
			},
		})
	}

	if dry {
		resp, err := c.Rig.Settings().GetSettings(ctx, connect.NewRequest(&settings_api.GetSettingsRequest{}))
		if err != nil {
			return err
		}

		s := resp.Msg.GetSettings()
		if err := utils.ApplySettingsUpdates(s, updates); err != nil {
			return err
		}

		t := flags.Flags.OutputType
		if t == common.OutputTypePretty {
			t = common.OutputTypeYAML
		}

		return common.FormatPrint(s, t)
	}

	if _, err := c.Rig.Settings().UpdateSettings(ctx, connect.NewRequest(&settings_api.UpdateSettingsRequest{
		Updates: updates,
	})); err != nil {
		return err
	}

	return nil
}
