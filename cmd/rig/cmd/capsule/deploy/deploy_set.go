package deploy

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) deploySet(ctx context.Context, cmd *cobra.Command, args []string) error {
	currentRolloutIDs, err := parseEnvRollouts(currentEnvRollouts)
	if err != nil {
		return err
	}

	capsule, err := c.getNewSpec(ctx, cmd, args)
	if err != nil {
		return err
	}

	changes := []*capsule_api.Change{{
		Field: &capsule_api.Change_Spec{
			Spec: capsule.GetSpec(),
		},
	}}

	respGit, err := c.Rig.Capsule().GetEffectiveGitSettings(
		ctx, connect.NewRequest(&capsule_api.GetEffectiveGitSettingsRequest{
			ProjectId: capsule.GetProject(),
			CapsuleId: capsule.GetName(),
		}),
	)
	if err != nil {
		return err
	}
	if respGit.Msg.GetGit().GetCapsuleSetPath() != "" && prBranchName != "" {
		resp, err := c.Rig.Capsule().ProposeSetRollout(ctx, connect.NewRequest(&capsule_api.ProposeSetRolloutRequest{
			CapsuleId:  capsule.GetName(),
			Changes:    changes,
			ProjectId:  capsule.GetProject(),
			BranchName: prBranchName,
		}))
		if err != nil {
			return err
		}
		url := resp.Msg.GetProposal().GetMetadata().GetReviewUrl()
		fmt.Println("New pull request created at", url)
		return nil
	} else if respGit.Msg.GetGit().GetCapsuleSetPath() == "" && prBranchName != "" {
		return errors.InvalidArgumentErrorf("--pr-branch was set, but the capsuleset is not git backed")
	}

	resp, err := c.Rig.Capsule().DeploySet(ctx, connect.NewRequest(&capsule_api.DeploySetRequest{
		CapsuleId:          capsule.GetName(),
		Changes:            changes,
		ProjectId:          capsule.GetProject(),
		CurrentRolloutIds:  currentRolloutIDs,
		CurrentFingerprint: parseFingerprint(currentFingerprint),
	}))
	if err != nil {
		return err
	}

	var inputs []capsule_cmd.WaitForRolloutInput
	for _, env := range resp.Msg.GetActiveEnvironments() {
		inputs = append(inputs, capsule_cmd.WaitForRolloutInput{
			RollbackInput: capsule_cmd.RollbackInput{
				BaseInput: capsule_cmd.BaseInput{
					Ctx:           ctx,
					Rig:           c.Rig,
					ProjectID:     capsule.GetProject(),
					EnvironmentID: env,
					CapsuleID:     capsule.GetName(),
				},
			},
			Fingerprints: &model.Fingerprints{
				CapsuleSet: resp.Msg.GetRevision().GetMetadata().GetFingerprint(),
			},
			PrintPrefix: env + "> ",
		})
	}

	if noWait {
		fmt.Println("Wrote capsule set change")
		return nil
	}

	fmt.Printf("Wrote capsule set change. Waiting for %v rollouts.\n", len(inputs))
	return waitForRollouts(inputs, timeout)
}

func parseEnvRollouts(s string) (map[string]uint64, error) {
	if s == "" {
		return nil, nil
	}

	res := map[string]uint64{}
	envs := strings.Split(s, ",")
	for _, env := range envs {
		ss := strings.Split(env, ":")
		if len(ss) != 2 {
			return nil, fmt.Errorf("malformed environment-rollout string: %s", s)
		}
		rolloutID, err := strconv.ParseUint(ss[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malfroemd environment-rollout string: %w", err)
		}
		res[ss[0]] = rolloutID
	}

	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func waitForRollouts(inputs []capsule_cmd.WaitForRolloutInput, timeout time.Duration) error {
	type inp struct {
		input capsule_cmd.WaitForRolloutInput
		state *capsule_cmd.WaitForRolloutState
	}
	var inps []inp
	for _, i := range inputs {
		i.PrintPrefix = i.EnvironmentID + "> "
		inps = append(inps, inp{
			input: i,
			state: &capsule_cmd.WaitForRolloutState{},
		})
	}
	start := time.Now()
	for len(inps) > 0 {
		if timeout > 0 && time.Since(start) > timeout {
			fmt.Println()
			fmt.Printf("🛑 Rollouts timed out after %s... ", timeout)
			return fmt.Errorf("aborted")
		}
		var finished []int
		for idx, input := range inps {
			if ok, err := capsule_cmd.WaitForRolloutIteration(input.input, input.state); err != nil {
				return err
			} else if ok {
				finished = append(finished, idx)
			}
		}
		finished = append(finished, len(inps)+1)
		f, i := 0, 0
		for idx := 0; idx < len(inps); idx++ {
			if idx == finished[f] {
				f++
			} else {
				inps[i] = inps[idx]
				i++
			}
		}
		inps = inps[:i]
		time.Sleep(time.Second)
	}

	return nil
}
