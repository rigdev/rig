package common

import (
	"strings"

	"github.com/spf13/cobra"
)

var BoolCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"true", "false"}, cobra.ShellCompDirectiveDefault
}

func MaxArgsCompletionFilter(max int) CompletionFilter {
	return func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
		current []string,
		directive cobra.ShellCompDirective,
	) ([]string, cobra.ShellCompDirective) {
		args = append(args, toComplete)
		if len(args) > max {
			return []string{}, cobra.ShellCompDirectiveError
		}
		return current, directive
	}
}

var ArgsCompletionFilter = func(
	cmd *cobra.Command,
	args []string,
	toComplete string,
	completions []string,
	directive cobra.ShellCompDirective,
) ([]string, cobra.ShellCompDirective) {
	args = append(args, toComplete)
	err := cmd.Args(cmd, args)
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError
	}

	return completions, directive
}

type CompletionFilter func(
	cmd *cobra.Command,
	args []string,
	toComplete string,
	current []string,
	directive cobra.ShellCompDirective,
) ([]string, cobra.ShellCompDirective)

type CompletionBase func(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective)

// Complete is a helper function to chain a completion function and subsequent filters.
func Complete(
	base CompletionBase,
	filters ...CompletionFilter,
) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	complete := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		completions, directive := base(cmd, args, toComplete)
		for _, f := range filters {
			completions, directive = f(cmd, args, toComplete, completions, directive)
			if directive == cobra.ShellCompDirectiveError {
				return []string{}, directive
			}
		}
		return completions, directive
	}

	return complete
}

func RoleCompletions(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	availableRoles := []string{"admin", "owner", "developer", "viewer"}

	var completions []string

	for _, r := range availableRoles {
		if strings.HasPrefix(r, toComplete) {
			completions = append(completions, r)
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveDefault
}
