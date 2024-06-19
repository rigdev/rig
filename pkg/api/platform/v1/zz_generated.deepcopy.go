//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Capsule) DeepCopyInto(out *Capsule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Capsule.
func (in *Capsule) DeepCopy() *Capsule {
	if in == nil {
		return nil
	}
	out := new(Capsule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Capsule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleSet) DeepCopyInto(out *CapsuleSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Spec.DeepCopyInto(&out.Spec)
	if in.Environments != nil {
		in, out := &in.Environments, &out.Environments
		*out = make(map[string]CapsuleSpec, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.EnvironmentRefs != nil {
		in, out := &in.EnvironmentRefs, &out.EnvironmentRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleSet.
func (in *CapsuleSet) DeepCopy() *CapsuleSet {
	if in == nil {
		return nil
	}
	out := new(CapsuleSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CapsuleSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleSpec) DeepCopyInto(out *CapsuleSpec) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make([]v1alpha2.CapsuleInterface, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Files != nil {
		in, out := &in.Files, &out.Files
		*out = make([]File, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Env.DeepCopyInto(&out.Env)
	in.Scale.DeepCopyInto(&out.Scale)
	if in.CronJobs != nil {
		in, out := &in.CronJobs, &out.CronJobs
		*out = make([]v1alpha2.CronJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleSpec.
func (in *CapsuleSpec) DeepCopy() *CapsuleSpec {
	if in == nil {
		return nil
	}
	out := new(CapsuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CapsuleSpec) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Environment) DeepCopyInto(out *Environment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Spec.DeepCopyInto(&out.Spec)
	if in.ActiveProjects != nil {
		in, out := &in.ActiveProjects, &out.ActiveProjects
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
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

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Environment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvironmentSource) DeepCopyInto(out *EnvironmentSource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvironmentSource.
func (in *EnvironmentSource) DeepCopy() *EnvironmentSource {
	if in == nil {
		return nil
	}
	out := new(EnvironmentSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvironmentVariables) DeepCopyInto(out *EnvironmentVariables) {
	*out = *in
	if in.Direct != nil {
		in, out := &in.Direct, &out.Direct
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Sources != nil {
		in, out := &in.Sources, &out.Sources
		*out = make([]EnvironmentSource, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvironmentVariables.
func (in *EnvironmentVariables) DeepCopy() *EnvironmentVariables {
	if in == nil {
		return nil
	}
	out := new(EnvironmentVariables)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *File) DeepCopyInto(out *File) {
	*out = *in
	if in.Bytes != nil {
		in, out := &in.Bytes, &out.Bytes
		*out = new([]byte)
		if **in != nil {
			in, out := *in, *out
			*out = make([]byte, len(*in))
			copy(*out, *in)
		}
	}
	if in.String != nil {
		in, out := &in.String, &out.String
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new File.
func (in *File) DeepCopy() *File {
	if in == nil {
		return nil
	}
	out := new(File)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostCapsule) DeepCopyInto(out *HostCapsule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Network.DeepCopyInto(&out.Network)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostCapsule.
func (in *HostCapsule) DeepCopy() *HostCapsule {
	if in == nil {
		return nil
	}
	out := new(HostCapsule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HostCapsule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostNetwork) DeepCopyInto(out *HostNetwork) {
	*out = *in
	if in.HostInterfaces != nil {
		in, out := &in.HostInterfaces, &out.HostInterfaces
		*out = make([]ProxyInterface, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.CapsuleInterface != nil {
		in, out := &in.CapsuleInterface, &out.CapsuleInterface
		*out = make([]ProxyInterface, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostNetwork.
func (in *HostNetwork) DeepCopy() *HostNetwork {
	if in == nil {
		return nil
	}
	out := new(HostNetwork)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InterfaceOptions) DeepCopyInto(out *InterfaceOptions) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InterfaceOptions.
func (in *InterfaceOptions) DeepCopy() *InterfaceOptions {
	if in == nil {
		return nil
	}
	out := new(InterfaceOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjEnvCapsuleBase) DeepCopyInto(out *ProjEnvCapsuleBase) {
	*out = *in
	if in.Files != nil {
		in, out := &in.Files, &out.Files
		*out = make([]File, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Env.DeepCopyInto(&out.Env)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjEnvCapsuleBase.
func (in *ProjEnvCapsuleBase) DeepCopy() *ProjEnvCapsuleBase {
	if in == nil {
		return nil
	}
	out := new(ProjEnvCapsuleBase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Project) DeepCopyInto(out *Project) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Project.
func (in *Project) DeepCopy() *Project {
	if in == nil {
		return nil
	}
	out := new(Project)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Project) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyInterface) DeepCopyInto(out *ProxyInterface) {
	*out = *in
	in.Options.DeepCopyInto(&out.Options)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyInterface.
func (in *ProxyInterface) DeepCopy() *ProxyInterface {
	if in == nil {
		return nil
	}
	out := new(ProxyInterface)
	in.DeepCopyInto(out)
	return out
}
