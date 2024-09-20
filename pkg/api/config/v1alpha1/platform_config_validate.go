package v1alpha1

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (cfg *PlatformConfig) Validate() error {
	if cfg.Cluster.Type != "" && len(cfg.Clusters) != 0 {
		return fmt.Errorf("only one of `cluster` and `clusters` must be set")
	}

	var errs field.ErrorList
	errs = append(errs, cfg.Cluster.validate(field.NewPath("clusters"))...)
	errs = append(errs, cfg.validateCapsuleExtensions(field.NewPath("capsuleExtensions"))...)
	errs = append(errs, cfg.Client.validate(field.NewPath("client"))...)

	return errs.ToAggregate()
}

func (c Cluster) validate(path *field.Path) field.ErrorList {
	return c.Git.validate(path.Child("git"))
}

func (g ClusterGit) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if g.PathPrefix != "" && g.PathPrefixes != (PathPrefixes{}) {
		return append(errs, field.Invalid(path, g, "can't set both `pathPrefix` and `pathPrefixes`"))
	}

	return errs
}

func (g Client) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, g.Git.validate(path.Child("git"))...)
	return errs
}

func (g ClientGit) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for idx, a := range g.Auths {
		errs = append(errs, a.validate(path.Index(idx))...)
	}
	for idx, g := range g.GitHubAuths {
		errs = append(errs, g.validate(path.Index(idx))...)
	}
	for idx, g := range g.GitLabAuths {
		errs = append(errs, g.validate(path.Index(idx))...)
	}
	return errs
}

func (a GitAuth) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if a.URL == "" && a.URLPrefix == "" {
		errs = append(errs, field.Invalid(path, a, "either `url` or `urlPrefix` must be set"))
	}
	if a.URL != "" && a.URLPrefix != "" {
		errs = append(errs, field.Invalid(path, a, "not both `url` and `urlPrefix` can be set"))
	}
	if a.URLPrefix != "" && a.Match != "" {
		errs = append(errs, field.Invalid(path, a, "not both `urlPrefix` and `match` can be given"))
	}
	if !slices.Contains([]string{"", "exact", "prefix"}, string(a.Match)) {
		errs = append(errs, field.Invalid(path.Child("match"), a.Match, "`match` must be one of '', 'exact' or 'prefix'"))
	}
	return errs
}

func (g GitHub) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if g.OrgRepo != "" && (g.Organization != "" || g.Repository != "") {
		errs = append(errs, field.Invalid(path, g, "not both 'orgRepo' and 'organization' or 'repository' can be given"))
	}
	if g.OrgRepo == "" && g.Organization == "" {
		errs = append(errs, field.Invalid(path.Child("orgRepo"), g.OrgRepo, "'orgRepo' must be given"))
	}

	if g.OrgRepo != "" {
		if !orgRepoRegex.MatchString(g.OrgRepo) {
			errs = append(errs, field.Invalid(path.Child("orgRepo"), g.OrgRepo, "`orgRepo` is malformed"))
		}
	}
	return errs
}

var orgRepoRegex = regexp.MustCompile("^[^/]+(/[^/]+)?$")

func (g GitLab) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if g.GroupsProject != "" && (len(g.Groups) != 0 || g.Project != "") {
		errs = append(errs, field.Invalid(path, g, "not both `groupsProject` and `groups` or `project` can be given"))
	}
	if g.GroupsProject == "" && len(g.Groups) == 0 {
		errs = append(errs, field.Invalid(path, g.GroupsProject, "`groupsProject` cannot be empty"))
	}
	if g.GroupsProject != "" {
		if !groupsProjectRegex.MatchString(g.GroupsProject) {
			errs = append(errs, field.Invalid(path, g.GroupsProject, "`groupsProject` is malformed"))
		}
	}
	return errs
}

var groupsProjectRegex = regexp.MustCompile("^[^/:]+(/[^/:]+)*(:[^/:]+)?$")

func (cfg *PlatformConfig) validateCapsuleExtensions(path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for k, v := range cfg.CapsuleExtensions {
		if err := v.validate(path.Key(k)); err != nil {
			errs = append(errs, err...)
		}
	}
	return errs
}

func (e Extension) validate(path *field.Path) field.ErrorList {
	var errs field.ErrorList

	schemaPath := path.Child("schema")
	if e.Schema == nil {
		errs = append(errs, field.Invalid(
			path, e, "value has no schema"),
		)
		return errs
	}
	if e.Schema.Type != "object" {
		errs = append(errs, field.Invalid(
			schemaPath.Child("type"), e.Schema.Type, "top level schema must be of type 'object'"),
		)
		return errs
	}

	for key, prop := range e.Schema.Properties {
		if prop.Type == "object" || prop.Type == "array" {
			errs = append(errs,
				field.Invalid(
					schemaPath.Child("properties").Key(key).Child("type"),
					prop.Type, "complex child properties are not yet supported"),
			)
		}
	}

	if _, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(e.Schema)); err != nil {
		errs = append(errs, field.Invalid(schemaPath, e.Schema, err.Error()))
	}

	return errs
}

func (cfg *PlatformConfig) Migrate() {
	for _, c := range cfg.Clusters {
		if c.Git.URL != "" && c.Git.Credentials != (GitCredentials{}) {
			cfg.Client.Git.Auths = append(cfg.Client.Git.Auths, GitAuth{
				URL:         c.Git.URL,
				Credentials: c.Git.Credentials,
			})
		}
		if cfg.Client.Git.Author.Name == "" {
			cfg.Client.Git.Author.Name = c.Git.Author.Name
			cfg.Client.Git.Author.Email = c.Git.Author.Email
		}
	}

	for idx, a := range cfg.Client.Git.Auths {
		if a.URLPrefix != "" {
			a.URL = a.URLPrefix
			a.Match = URLMatchPrefix
			a.URLPrefix = ""
		}
		if a.Match == "" {
			a.Match = URLMatchExact
		}
		cfg.Client.Git.Auths[idx] = a
	}

	for idx, g := range cfg.Client.Git.GitHubAuths {
		if g.Organization != "" {
			g.OrgRepo = g.Organization
			if g.Repository != "" {
				g.OrgRepo += "/" + g.Repository
			}
		}
		g.Organization = ""
		g.Repository = ""
		cfg.Client.Git.GitHubAuths[idx] = g
	}

	for idx, g := range cfg.Client.Git.GitLabAuths {
		if len(g.Groups) > 0 {
			g.GroupsProject = strings.Join(g.Groups, "/")
		}
		if g.Project != "" {
			g.GroupsProject += ":" + g.Project
		}
		g.Groups = nil
		g.Project = ""
		cfg.Client.Git.GitLabAuths[idx] = g
	}

	if cfg.Client.Mailjet != (ClientMailjet{}) {
		cfg.Client.Mailjets["mailjet"] = cfg.Client.Mailjet
	}

	if cfg.Client.SMTP != (ClientSMTP{}) {
		cfg.Client.SMTPs["smtp"] = cfg.Client.SMTP
	}

	if cfg.Email != (Email{}) && cfg.Email.Type != "" {
		switch cfg.Email.Type {
		case EmailTypeMailjet:
			if cfg.Email.ID == "" {
				cfg.Email.ID = "mailjet"
			}
		case EmailTypeSMTP:
			if cfg.Email.ID == "" {
				cfg.Email.ID = "smtp"
			}
		}
	}
}
