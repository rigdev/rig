package cmd

import (
	"context"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	auth_service "github.com/rigdev/rig/internal/service/auth"
	project_service "github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var projectName string

func init() {
	projects := &cobra.Command{
		Use: "projects",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: register(ProjectsCreate),
	}
	create.PersistentFlags().StringVar(&projectName, "name", "", "name of the project")
	projects.AddCommand(create)

	delete := &cobra.Command{
		Use:  "delete <project-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(ProjectsDelete),
	}
	projects.AddCommand(delete)

	list := &cobra.Command{
		Use:  "list",
		RunE: register(ProjectsList),
	}
	projects.AddCommand(list)

	use := &cobra.Command{
		Use:  "use <project-id> <user-id>",
		Args: cobra.ExactArgs(2),
		RunE: register(ProjectsUse),
	}
	projects.AddCommand(use)

	rootCmd.AddCommand(projects)
}

func ProjectsCreate(ctx context.Context, cmd *cobra.Command, ps project_service.Service, logger *zap.Logger) error {
	var ups []*project.Update

	if projectName != "" {
		ups = append(ups, &project.Update{Field: &project.Update_Name{Name: projectName}})
	}

	p, err := ps.CreateProject(ctx, ups)
	if err != nil {
		return err
	}

	logger.Info("created project", zap.String("project_id", p.GetProjectId()), zap.String("name", p.GetName()))

	return nil
}

func ProjectsDelete(ctx context.Context, cmd *cobra.Command, args []string, ps project_service.Service, logger *zap.Logger) error {
	projectID := args[0]
	if projectID == "rig" {
		projectID = auth.RigProjectID
	}

	ctx = auth.WithProjectID(ctx, projectID)
	if err := ps.DeleteProject(ctx); err != nil {
		return err
	}

	logger.Info("deleted project", zap.String("project_id", projectID))

	return nil
}

func ProjectsList(ctx context.Context, cmd *cobra.Command, ps project_service.Service, logger *zap.Logger) error {
	it, _, err := ps.List(ctx, &model.Pagination{})
	if err != nil {
		return err
	}

	for {
		p, err := it.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		logger.Info("found project", zap.String("name", p.GetName()), zap.String("project_id", p.GetProjectId()))
	}
}

func ProjectsUse(ctx context.Context, cmd *cobra.Command, args []string, as *auth_service.Service, logger *zap.Logger) error {
	projectID := args[0]
	if projectID == "rig" {
		projectID = auth.RigProjectID
	}

	userID, err := uuid.Parse(args[1])
	if err != nil {
		return err
	}

	ctx = auth.WithClaims(ctx, auth_service.RigClaims{
		ProjectID:   auth.RigProjectID,
		Subject:     userID,
		SubjectType: auth.SubjectTypeUser,
	})
	t, err := as.UseProject(ctx, projectID)
	if err != nil {
		return err
	}

	logger.Info("using project", zap.String("project_id", projectID), zap.String("project_token", t))

	return nil
}
