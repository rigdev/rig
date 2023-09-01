package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func ProjectUse(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *base.Config, logger *zap.Logger) error {
	var projectID uuid.UUID
	var err error
	if len(args) != 1 {
		res, err := nc.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return err
		}

		var ps []string
		for _, p := range res.Msg.GetProjects() {
			ps = append(ps, p.GetName())
		}

		i, _, err := utils.PromptSelect("Project: ", ps, false)
		if err != nil {
			return err
		}

		projectID, err = uuid.Parse(res.Msg.GetProjects()[i].GetProjectId())
		if err != nil {
			return err
		}
	} else {
		if id, err := uuid.Parse(args[0]); err == nil {
			projectID = id
		} else {
			res, err := nc.Project().List(ctx, &connect.Request[project.ListRequest]{})
			if err != nil {
				return err
			}

			for _, p := range res.Msg.GetProjects() {
				if p.GetName() == args[0] {
					projectID, err = uuid.Parse(p.GetProjectId())
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	if projectID == uuid.Nil {
		return errors.NotFoundErrorf("project '%v' not found", args[0])
	}

	res, err := nc.Project().Use(ctx, &connect.Request[project.UseRequest]{
		Msg: &project.UseRequest{
			ProjectId: projectID.String(),
		},
	})
	if err != nil {
		return err
	}

	cfg.Context().Project.ProjectID = projectID
	cfg.Context().Project.ProjectToken = res.Msg.GetProjectToken()
	if err := cfg.Save(); err != nil {
		return err
	}

	cmd.Println("Changed project successfully!")

	return nil
}
