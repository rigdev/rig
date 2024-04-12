//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Auth) DeepCopyInto(out *Auth) {
	*out = *in
	in.SSO.DeepCopyInto(&out.SSO)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Auth.
func (in *Auth) DeepCopy() *Auth {
	if in == nil {
		return nil
	}
	out := new(Auth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleMatch) DeepCopyInto(out *CapsuleMatch) {
	*out = *in
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Names != nil {
		in, out := &in.Names, &out.Names
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleMatch.
func (in *CapsuleMatch) DeepCopy() *CapsuleMatch {
	if in == nil {
		return nil
	}
	out := new(CapsuleMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Client) DeepCopyInto(out *Client) {
	*out = *in
	out.Postgres = in.Postgres
	out.Docker = in.Docker
	out.Mailjet = in.Mailjet
	out.SMTP = in.SMTP
	out.Operator = in.Operator
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Client.
func (in *Client) DeepCopy() *Client {
	if in == nil {
		return nil
	}
	out := new(Client)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientDocker) DeepCopyInto(out *ClientDocker) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientDocker.
func (in *ClientDocker) DeepCopy() *ClientDocker {
	if in == nil {
		return nil
	}
	out := new(ClientDocker)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientMailjet) DeepCopyInto(out *ClientMailjet) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientMailjet.
func (in *ClientMailjet) DeepCopy() *ClientMailjet {
	if in == nil {
		return nil
	}
	out := new(ClientMailjet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientOperator) DeepCopyInto(out *ClientOperator) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientOperator.
func (in *ClientOperator) DeepCopy() *ClientOperator {
	if in == nil {
		return nil
	}
	out := new(ClientOperator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientPostgres) DeepCopyInto(out *ClientPostgres) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientPostgres.
func (in *ClientPostgres) DeepCopy() *ClientPostgres {
	if in == nil {
		return nil
	}
	out := new(ClientPostgres)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientSMTP) DeepCopyInto(out *ClientSMTP) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientSMTP.
func (in *ClientSMTP) DeepCopy() *ClientSMTP {
	if in == nil {
		return nil
	}
	out := new(ClientSMTP)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.DevRegistry = in.DevRegistry
	out.Git = in.Git
	if in.CreatePullSecrets != nil {
		in, out := &in.CreatePullSecrets, &out.CreatePullSecrets
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterGit) DeepCopyInto(out *ClusterGit) {
	*out = *in
	out.PathPrefixes = in.PathPrefixes
	out.Credentials = in.Credentials
	out.Author = in.Author
	out.Templates = in.Templates
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterGit.
func (in *ClusterGit) DeepCopy() *ClusterGit {
	if in == nil {
		return nil
	}
	out := new(ClusterGit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomPlugin) DeepCopyInto(out *CustomPlugin) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomPlugin.
func (in *CustomPlugin) DeepCopy() *CustomPlugin {
	if in == nil {
		return nil
	}
	out := new(CustomPlugin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevRegistry) DeepCopyInto(out *DevRegistry) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevRegistry.
func (in *DevRegistry) DeepCopy() *DevRegistry {
	if in == nil {
		return nil
	}
	out := new(DevRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Email) DeepCopyInto(out *Email) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Email.
func (in *Email) DeepCopy() *Email {
	if in == nil {
		return nil
	}
	out := new(Email)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Environment) DeepCopyInto(out *Environment) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Environment.
func (in *Environment) DeepCopy() *Environment {
	if in == nil {
		return nil
	}
	out := new(Environment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitAuthor) DeepCopyInto(out *GitAuthor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitAuthor.
func (in *GitAuthor) DeepCopy() *GitAuthor {
	if in == nil {
		return nil
	}
	out := new(GitAuthor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitCredentials) DeepCopyInto(out *GitCredentials) {
	*out = *in
	out.HTTPS = in.HTTPS
	out.SSH = in.SSH
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitCredentials.
func (in *GitCredentials) DeepCopy() *GitCredentials {
	if in == nil {
		return nil
	}
	out := new(GitCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitTemplates) DeepCopyInto(out *GitTemplates) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitTemplates.
func (in *GitTemplates) DeepCopy() *GitTemplates {
	if in == nil {
		return nil
	}
	out := new(GitTemplates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPSCredential) DeepCopyInto(out *HTTPSCredential) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPSCredential.
func (in *HTTPSCredential) DeepCopy() *HTTPSCredential {
	if in == nil {
		return nil
	}
	out := new(HTTPSCredential)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Logging) DeepCopyInto(out *Logging) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Logging.
func (in *Logging) DeepCopy() *Logging {
	if in == nil {
		return nil
	}
	out := new(Logging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCProvider) DeepCopyInto(out *OIDCProvider) {
	*out = *in
	if in.AllowedDomains != nil {
		in, out := &in.AllowedDomains, &out.AllowedDomains
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DisableJITGroups != nil {
		in, out := &in.DisableJITGroups, &out.DisableJITGroups
		*out = new(bool)
		**out = **in
	}
	if in.GroupMapping != nil {
		in, out := &in.GroupMapping, &out.GroupMapping
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.DisableUserMerging != nil {
		in, out := &in.DisableUserMerging, &out.DisableUserMerging
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCProvider.
func (in *OIDCProvider) DeepCopy() *OIDCProvider {
	if in == nil {
		return nil
	}
	out := new(OIDCProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperatorConfig) DeepCopyInto(out *OperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.WebhooksEnabled != nil {
		in, out := &in.WebhooksEnabled, &out.WebhooksEnabled
		*out = new(bool)
		**out = **in
	}
	if in.LeaderElectionEnabled != nil {
		in, out := &in.LeaderElectionEnabled, &out.LeaderElectionEnabled
		*out = new(bool)
		**out = **in
	}
	if in.PrometheusServiceMonitor != nil {
		in, out := &in.PrometheusServiceMonitor, &out.PrometheusServiceMonitor
		*out = new(PrometheusServiceMonitor)
		**out = **in
	}
	out.VerticalPodAutoscaler = in.VerticalPodAutoscaler
	in.Pipeline.DeepCopyInto(&out.Pipeline)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperatorConfig.
func (in *OperatorConfig) DeepCopy() *OperatorConfig {
	if in == nil {
		return nil
	}
	out := new(OperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OperatorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PathPrefixes) DeepCopyInto(out *PathPrefixes) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PathPrefixes.
func (in *PathPrefixes) DeepCopy() *PathPrefixes {
	if in == nil {
		return nil
	}
	out := new(PathPrefixes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Pipeline) DeepCopyInto(out *Pipeline) {
	*out = *in
	out.RoutesStep = in.RoutesStep
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make([]Step, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CustomPlugins != nil {
		in, out := &in.CustomPlugins, &out.CustomPlugins
		*out = make([]CustomPlugin, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Pipeline.
func (in *Pipeline) DeepCopy() *Pipeline {
	if in == nil {
		return nil
	}
	out := new(Pipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PlatformConfig) DeepCopyInto(out *PlatformConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Auth.DeepCopyInto(&out.Auth)
	out.Client = in.Client
	out.Repository = in.Repository
	in.Cluster.DeepCopyInto(&out.Cluster)
	out.Email = in.Email
	out.Logging = in.Logging
	if in.Clusters != nil {
		in, out := &in.Clusters, &out.Clusters
		*out = make(map[string]Cluster, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Environments != nil {
		in, out := &in.Environments, &out.Environments
		*out = make(map[string]Environment, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PlatformConfig.
func (in *PlatformConfig) DeepCopy() *PlatformConfig {
	if in == nil {
		return nil
	}
	out := new(PlatformConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PlatformConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Plugin) DeepCopyInto(out *Plugin) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Plugin.
func (in *Plugin) DeepCopy() *Plugin {
	if in == nil {
		return nil
	}
	out := new(Plugin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrometheusServiceMonitor) DeepCopyInto(out *PrometheusServiceMonitor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrometheusServiceMonitor.
func (in *PrometheusServiceMonitor) DeepCopy() *PrometheusServiceMonitor {
	if in == nil {
		return nil
	}
	out := new(PrometheusServiceMonitor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Repository) DeepCopyInto(out *Repository) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Repository.
func (in *Repository) DeepCopy() *Repository {
	if in == nil {
		return nil
	}
	out := new(Repository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutesStep) DeepCopyInto(out *RoutesStep) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutesStep.
func (in *RoutesStep) DeepCopy() *RoutesStep {
	if in == nil {
		return nil
	}
	out := new(RoutesStep)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSHCredential) DeepCopyInto(out *SSHCredential) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSHCredential.
func (in *SSHCredential) DeepCopy() *SSHCredential {
	if in == nil {
		return nil
	}
	out := new(SSHCredential)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SSO) DeepCopyInto(out *SSO) {
	*out = *in
	if in.OIDCProviders != nil {
		in, out := &in.OIDCProviders, &out.OIDCProviders
		*out = make(map[string]OIDCProvider, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SSO.
func (in *SSO) DeepCopy() *SSO {
	if in == nil {
		return nil
	}
	out := new(SSO)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Step) DeepCopyInto(out *Step) {
	*out = *in
	in.Match.DeepCopyInto(&out.Match)
	if in.Plugins != nil {
		in, out := &in.Plugins, &out.Plugins
		*out = make([]Plugin, len(*in))
		copy(*out, *in)
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Capsules != nil {
		in, out := &in.Capsules, &out.Capsules
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Step.
func (in *Step) DeepCopy() *Step {
	if in == nil {
		return nil
	}
	out := new(Step)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VerticalPodAutoscaler) DeepCopyInto(out *VerticalPodAutoscaler) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VerticalPodAutoscaler.
func (in *VerticalPodAutoscaler) DeepCopy() *VerticalPodAutoscaler {
	if in == nil {
		return nil
	}
	out := new(VerticalPodAutoscaler)
	in.DeepCopyInto(out)
	return out
}
