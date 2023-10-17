package scale

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) horizontal(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return nil
	}

	horizontal := rollout.GetConfig().GetHorizontalScale()
	if horizontal == nil {
		horizontal = &capsule.HorizontalScale{}
	}

	if horizontal.CpuTarget != nil && !overwriteAutoscaler {
		return errors.New("cannot set the number of replicas with the autoscaler enabled with setting the --overwrite-autoscaler flag")
	}

	horizontal.CpuTarget = nil

	if !cmd.Flags().Lookup("replicas").Changed {
		return errors.New("--replicas not set")
	}
	horizontal.MinReplicas = replicas
	horizontal.MaxReplicas = replicas

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_HorizontalScale{
					HorizontalScale: horizontal,
				},
			},
		},
	})

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.AbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}

	return nil
}

func (c Cmd) autoscale(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return nil
	}
	replicas := rollout.GetConfig().GetReplicas()
	horizontal := rollout.GetConfig().GetHorizontalScale()
	if horizontal == nil {
		horizontal = &capsule.HorizontalScale{}
	}

	if cmd.Flags().Lookup("min-replicas").Changed {
		horizontal.MinReplicas = uint32(minReplicas)
	}
	if cmd.Flags().Lookup("max-replicas").Changed {
		horizontal.MaxReplicas = uint32(maxReplicas)
	}
	if cmd.Flags().Lookup("utilization-percentage").Changed {
		cpuTarget := horizontal.GetCpuTarget()
		if cpuTarget == nil {
			cpuTarget = &capsule.CPUTarget{}
		}
		cpuTarget.AverageUtilizationPercentage = uint32(utilizationPercentage)
		horizontal.CpuTarget = cpuTarget
	}

	if disable {
		horizontal.CpuTarget = nil
		horizontal.MinReplicas = replicas
		horizontal.MaxReplicas = replicas
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_HorizontalScale{
					HorizontalScale: horizontal,
				},
			},
		},
	})

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.AbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}

	return nil
}
