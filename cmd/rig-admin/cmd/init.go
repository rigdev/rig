package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"

	auth_service "github.com/rigdev/rig/internal/service/auth"
	project_service "github.com/rigdev/rig/internal/service/project"
	user_service "github.com/rigdev/rig/internal/service/user"
	"github.com/spf13/cobra"
)

func init() {
	init := &cobra.Command{
		Use:   "init",
		RunE:  register(initialize),
		Short: "Sets up the Rig server",
	}
	rootCmd.AddCommand(init)
}

func initialize(ctx context.Context, cmd *cobra.Command, us user_service.Service, ps project_service.Service) error {
	printBanner("Create user")
	if err := createUser(ctx, us); err != nil {
		return err
	}
	fmt.Println()

	printBanner("Project")
	if err := createProject(ctx, ps); err != nil {
		return err
	}
	fmt.Println()

	fmt.Println("Rig is set setup :)")

	return nil
}

func createUser(ctx context.Context, us user_service.Service) error {
	_, total, err := us.List(ctx, &model.Pagination{
		Limit: 1,
	}, "")
	if err != nil {
		return err
	}

	shouldCreateNew := false
	if total == 0 {
		fmt.Println("There are no admin users, create a new one")
		shouldCreateNew = true
	} else {
		pluralS := "s"
		if total == 1 {
			pluralS = ""
		}
		fmt.Printf("There are already %v admin user%s.\n", total, pluralS)
		shouldCreateNew, err = common.PromptConfirm("Do you wish to create a new one anyways?", false)
		if err != nil {
			return err
		}
	}

	if !shouldCreateNew {
		return nil
	}

	email, err := common.PromptGetInput("Email of the new user:", common.ValidateEmail)
	if err != nil {
		return err
	}
	password, err := common.GetPasswordPrompt("Password:")
	if err != nil {
		return err
	}

	ctx = auth.WithClaims(ctx, auth_service.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})
	_, err = us.CreateUser(ctx, &model.RegisterMethod{
		Method: &model.RegisterMethod_System_{},
	}, []*user.Update{
		{Field: &user.Update_Email{Email: email}},
		{Field: &user.Update_Password{Password: password}},
	})
	if err != nil {
		return err
	}
	fmt.Println("User created successfully!")

	return nil
}

func createProject(ctx context.Context, ps project_service.Service) error {
	_, total, err := ps.List(ctx, &model.Pagination{
		Limit: 1,
	})
	if err != nil {
		return err
	}

	if total > 0 {
		msg := fmt.Sprintf("There already exist %v projects. Do you want to create a new one anyways?", total)
		use, err := common.PromptConfirm(msg, false)
		if err != nil {
			return err
		}
		if !use {
			return nil
		}
	} else {
		fmt.Println("There are no projects, create a new one.")
	}

	for {
		projectName, err := common.PromptGetInput("Project name:", common.ValidateNonEmpty)
		if err != nil {
			return err
		}

		_, err = ps.CreateProject(ctx, []*project.Update{{Field: &project.Update_Name{Name: projectName}}})
		if errors.IsAlreadyExists(err) {
			fmt.Println("A project with that name already exists")
			continue
		}
		if err != nil {
			return err
		}
		break
	}
	fmt.Println("Project created succesfully!")

	return nil
}

func printBanner(s string) {
	color.Cyan(makeBanner(s))
}
func makeBanner(s string) string {
	totalLength := 60
	sideLength := (totalLength - len(s)) / 2
	side := strings.Repeat("-", sideLength)
	return side + " " + s + " " + side
}
