//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CPUTarget) DeepCopyInto(out *CPUTarget) {
	*out = *in
	if in.Utilization != nil {
		in, out := &in.Utilization, &out.Utilization
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CPUTarget.
func (in *CPUTarget) DeepCopy() *CPUTarget {
	if in == nil {
		return nil
	}
	out := new(CPUTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Capsule) DeepCopyInto(out *Capsule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(CapsuleStatus)
		(*in).DeepCopyInto(*out)
	}
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
func (in *CapsuleInterface) DeepCopyInto(out *CapsuleInterface) {
	*out = *in
	if in.Liveness != nil {
		in, out := &in.Liveness, &out.Liveness
		*out = new(InterfaceProbe)
		(*in).DeepCopyInto(*out)
	}
	if in.Readiness != nil {
		in, out := &in.Readiness, &out.Readiness
		*out = new(InterfaceProbe)
		(*in).DeepCopyInto(*out)
	}
	if in.Public != nil {
		in, out := &in.Public, &out.Public
		*out = new(CapsulePublicInterface)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleInterface.
func (in *CapsuleInterface) DeepCopy() *CapsuleInterface {
	if in == nil {
		return nil
	}
	out := new(CapsuleInterface)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleInterfaceIngress) DeepCopyInto(out *CapsuleInterfaceIngress) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleInterfaceIngress.
func (in *CapsuleInterfaceIngress) DeepCopy() *CapsuleInterfaceIngress {
	if in == nil {
		return nil
	}
	out := new(CapsuleInterfaceIngress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleInterfaceLoadBalancer) DeepCopyInto(out *CapsuleInterfaceLoadBalancer) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleInterfaceLoadBalancer.
func (in *CapsuleInterfaceLoadBalancer) DeepCopy() *CapsuleInterfaceLoadBalancer {
	if in == nil {
		return nil
	}
	out := new(CapsuleInterfaceLoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleList) DeepCopyInto(out *CapsuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Capsule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleList.
func (in *CapsuleList) DeepCopy() *CapsuleList {
	if in == nil {
		return nil
	}
	out := new(CapsuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CapsuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsulePublicInterface) DeepCopyInto(out *CapsulePublicInterface) {
	*out = *in
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(CapsuleInterfaceIngress)
		**out = **in
	}
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(CapsuleInterfaceLoadBalancer)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsulePublicInterface.
func (in *CapsulePublicInterface) DeepCopy() *CapsulePublicInterface {
	if in == nil {
		return nil
	}
	out := new(CapsulePublicInterface)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleScale) DeepCopyInto(out *CapsuleScale) {
	*out = *in
	in.Horizontal.DeepCopyInto(&out.Horizontal)
	if in.Vertical != nil {
		in, out := &in.Vertical, &out.Vertical
		*out = new(VerticalScale)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleScale.
func (in *CapsuleScale) DeepCopy() *CapsuleScale {
	if in == nil {
		return nil
	}
	out := new(CapsuleScale)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleSpec) DeepCopyInto(out *CapsuleSpec) {
	*out = *in
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make([]CapsuleInterface, len(*in))
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
	in.Scale.DeepCopyInto(&out.Scale)
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = new(Env)
		(*in).DeepCopyInto(*out)
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CapsuleStatus) DeepCopyInto(out *CapsuleStatus) {
	*out = *in
	if in.OwnedResources != nil {
		in, out := &in.OwnedResources, &out.OwnedResources
		*out = make([]OwnedResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.UsedResources != nil {
		in, out := &in.UsedResources, &out.UsedResources
		*out = make([]UsedResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(DeploymentStatus)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CapsuleStatus.
func (in *CapsuleStatus) DeepCopy() *CapsuleStatus {
	if in == nil {
		return nil
	}
	out := new(CapsuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomMetric) DeepCopyInto(out *CustomMetric) {
	*out = *in
	if in.InstanceMetric != nil {
		in, out := &in.InstanceMetric, &out.InstanceMetric
		*out = new(InstanceMetric)
		(*in).DeepCopyInto(*out)
	}
	if in.ObjectMetric != nil {
		in, out := &in.ObjectMetric, &out.ObjectMetric
		*out = new(ObjectMetric)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomMetric.
func (in *CustomMetric) DeepCopy() *CustomMetric {
	if in == nil {
		return nil
	}
	out := new(CustomMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentStatus) DeepCopyInto(out *DeploymentStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentStatus.
func (in *DeploymentStatus) DeepCopy() *DeploymentStatus {
	if in == nil {
		return nil
	}
	out := new(DeploymentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Env) DeepCopyInto(out *Env) {
	*out = *in
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = make([]EnvReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Env.
func (in *Env) DeepCopy() *Env {
	if in == nil {
		return nil
	}
	out := new(Env)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvReference) DeepCopyInto(out *EnvReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvReference.
func (in *EnvReference) DeepCopy() *EnvReference {
	if in == nil {
		return nil
	}
	out := new(EnvReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *File) DeepCopyInto(out *File) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(FileContentReference)
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
func (in *FileContentReference) DeepCopyInto(out *FileContentReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileContentReference.
func (in *FileContentReference) DeepCopy() *FileContentReference {
	if in == nil {
		return nil
	}
	out := new(FileContentReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalScale) DeepCopyInto(out *HorizontalScale) {
	*out = *in
	in.Instances.DeepCopyInto(&out.Instances)
	if in.CPUTarget != nil {
		in, out := &in.CPUTarget, &out.CPUTarget
		*out = new(CPUTarget)
		(*in).DeepCopyInto(*out)
	}
	if in.CustomMetrics != nil {
		in, out := &in.CustomMetrics, &out.CustomMetrics
		*out = make([]CustomMetric, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalScale.
func (in *HorizontalScale) DeepCopy() *HorizontalScale {
	if in == nil {
		return nil
	}
	out := new(HorizontalScale)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstanceMetric) DeepCopyInto(out *InstanceMetric) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstanceMetric.
func (in *InstanceMetric) DeepCopy() *InstanceMetric {
	if in == nil {
		return nil
	}
	out := new(InstanceMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Instances) DeepCopyInto(out *Instances) {
	*out = *in
	if in.Max != nil {
		in, out := &in.Max, &out.Max
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Instances.
func (in *Instances) DeepCopy() *Instances {
	if in == nil {
		return nil
	}
	out := new(Instances)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InterfaceGRPCProbe) DeepCopyInto(out *InterfaceGRPCProbe) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InterfaceGRPCProbe.
func (in *InterfaceGRPCProbe) DeepCopy() *InterfaceGRPCProbe {
	if in == nil {
		return nil
	}
	out := new(InterfaceGRPCProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InterfaceProbe) DeepCopyInto(out *InterfaceProbe) {
	*out = *in
	if in.GRPC != nil {
		in, out := &in.GRPC, &out.GRPC
		*out = new(InterfaceGRPCProbe)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InterfaceProbe.
func (in *InterfaceProbe) DeepCopy() *InterfaceProbe {
	if in == nil {
		return nil
	}
	out := new(InterfaceProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectMetric) DeepCopyInto(out *ObjectMetric) {
	*out = *in
	if in.MatchLabels != nil {
		in, out := &in.MatchLabels, &out.MatchLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.DescribedObject = in.DescribedObject
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectMetric.
func (in *ObjectMetric) DeepCopy() *ObjectMetric {
	if in == nil {
		return nil
	}
	out := new(ObjectMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OwnedResource) DeepCopyInto(out *OwnedResource) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(v1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OwnedResource.
func (in *OwnedResource) DeepCopy() *OwnedResource {
	if in == nil {
		return nil
	}
	out := new(OwnedResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceLimits) DeepCopyInto(out *ResourceLimits) {
	*out = *in
	if in.Request != nil {
		in, out := &in.Request, &out.Request
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.Limit != nil {
		in, out := &in.Limit, &out.Limit
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceLimits.
func (in *ResourceLimits) DeepCopy() *ResourceLimits {
	if in == nil {
		return nil
	}
	out := new(ResourceLimits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceRequest) DeepCopyInto(out *ResourceRequest) {
	*out = *in
	out.Request = in.Request.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceRequest.
func (in *ResourceRequest) DeepCopy() *ResourceRequest {
	if in == nil {
		return nil
	}
	out := new(ResourceRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UsedResource) DeepCopyInto(out *UsedResource) {
	*out = *in
	if in.Ref != nil {
		in, out := &in.Ref, &out.Ref
		*out = new(v1.TypedLocalObjectReference)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UsedResource.
func (in *UsedResource) DeepCopy() *UsedResource {
	if in == nil {
		return nil
	}
	out := new(UsedResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VerticalScale) DeepCopyInto(out *VerticalScale) {
	*out = *in
	if in.CPU != nil {
		in, out := &in.CPU, &out.CPU
		*out = new(ResourceLimits)
		(*in).DeepCopyInto(*out)
	}
	if in.Memory != nil {
		in, out := &in.Memory, &out.Memory
		*out = new(ResourceLimits)
		(*in).DeepCopyInto(*out)
	}
	if in.GPU != nil {
		in, out := &in.GPU, &out.GPU
		*out = new(ResourceRequest)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VerticalScale.
func (in *VerticalScale) DeepCopy() *VerticalScale {
	if in == nil {
		return nil
	}
	out := new(VerticalScale)
	in.DeepCopyInto(out)
	return out
}
