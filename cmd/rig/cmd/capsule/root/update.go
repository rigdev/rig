package root

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) update(ctx context.Context, _ *cobra.Command, _ []string) error {
	if !c.Scope.IsInteractive() {
		return errors.New("non-interactive mode is not supported for this command")
	}
	capsuleID := capsule_cmd.CapsuleID
	resp, err := c.Rig.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: capsuleID,
		ProjectId: flags.GetProject(c.Scope),
	}))
	if err != nil {
		return err
	}
	cc := resp.Msg.GetCapsule()

	envResp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{
		ProjectFilter: flags.GetProject(c.Scope),
	}))
	if err != nil {
		return err
	}

	var updates []*capsule.Update
	for {
		i, _, err := c.Prompter.Select("Select the setting to update (CTRL + C to cancel)",
			[]string{"Git store", "Done"})
		if err != nil {
			if common.ErrIsAborted(err) {
				return nil
			}
			return err
		}

		done := false
		switch i {
		case 0:
			cc.GitStore, err = common.PromptGitStore(c.Prompter, cc.GetGitStore(), envResp.Msg.GetEnvironments())
			if err != nil {
				if common.ErrIsAborted(err) {
					continue
				}
				return err
			}

			updates = append(updates, &capsule.Update{
				Field: &capsule.Update_SetGitStore{
					SetGitStore: cc.GitStore,
				},
			})
		case 2:
			done = true
		}
		if done {
			break
		}
	}

	if len(updates) == 0 {
		fmt.Println("No updates to make")
		return nil
	}

	if _, err = c.Rig.Capsule().Update(ctx, connect.NewRequest(&capsule.UpdateRequest{
		Updates:   updates,
		ProjectId: flags.GetProject(c.Scope),
		CapsuleId: capsuleID,
	})); err != nil {
		return err
	}

	fmt.Println("Capsule updated")

	return nil
}
