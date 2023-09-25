package auth

import (
	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c Cmd) getAuthConfig(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var projectID uuid.UUID
	var err error
	if len(args) != 1 {
		res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return err
		}

		ps := []string{"Rig"}
		psIds := []string{auth.RigProjectID.String()}
		for _, p := range res.Msg.GetProjects() {
			ps = append(ps, p.GetName())
			psIds = append(psIds, p.GetProjectId())
		}

		i, _, err := common.PromptSelect("Project: ", ps)
		if err != nil {
			return err
		}

		projectID, err = uuid.Parse(psIds[i])
		if err != nil {
			return err
		}
	} else {
		if id, err := uuid.Parse(args[0]); err == nil {
			projectID = id
		} else {
			res, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
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

	if redirectAddr == "" {
		redirectAddr, err = common.PromptInput("Oauth Redirect Address", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	res, err := c.Rig.Authentication().GetAuthConfig(ctx, &connect.Request[authentication.GetAuthConfigRequest]{
		Msg: &authentication.GetAuthConfigRequest{
			RedirectAddr: redirectAddr,
			ProjectId:    projectID.String(),
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		cmd.Println(common.ProtoToPrettyJson(res.Msg))
		return nil
	}

	rows_login := []table.Row{}
	for i, l := range res.Msg.GetLoginTypes() {
		if i == 0 {
			rows_login = append(rows_login, table.Row{"Login Mechanisms", l})
			continue
		}
		rows_login = append(rows_login, table.Row{"", l})
	}

	oauthSettings := res.Msg.GetOauthProviders()
	rows_oauth := []table.Row{}
	for _, o := range oauthSettings {
		rows_oauth = append(rows_oauth, table.Row{o.GetName(), o.GetProviderUrl()})
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"Allow Register", res.Msg.AllowsRegister},
	})
	for _, r := range rows_login {
		t.AppendRow(r)
	}

	cmd.Println(t.Render())
	for _, r := range rows_oauth {
		cmd.Println(r[0], ": ", r[1])
	}
	return nil
}
