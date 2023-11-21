package v1alpha1

import (
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/ptr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Capsule{}

func (src *Capsule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Capsule)

	srcSpec := src.DeepCopy().Spec

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Args = srcSpec.Args
	dst.Spec.Command = srcSpec.Command
	dst.Spec.Image = srcSpec.Image
	dst.Spec.NodeSelector = srcSpec.NodeSelector

	for _, f := range srcSpec.Files {
		switch {
		case f.ConfigMap != nil:
			dst.Spec.Files = append(dst.Spec.Files, v1alpha2.File{
				Ref: &v1alpha2.FileContentReference{
					Kind: "ConfigMap",
					Name: f.ConfigMap.Name,
					Key:  f.ConfigMap.Key,
				},
				Path: f.Path,
			})
		case f.Secret != nil:
			dst.Spec.Files = append(dst.Spec.Files, v1alpha2.File{
				Ref: &v1alpha2.FileContentReference{
					Kind: "Secret",
					Name: f.Secret.Name,
					Key:  f.Secret.Key,
				},
				Path: f.Path,
			})
		}
	}

	for _, i := range srcSpec.Interfaces {
		ni := v1alpha2.CapsuleInterface{
			Name: i.Name,
			Port: i.Port,
		}
		if i.Public != nil {
			ni.Public = &v1alpha2.CapsulePublicInterface{}
			if i.Public.Ingress != nil {
				ni.Public.Ingress = &v1alpha2.CapsuleInterfaceIngress{
					Host: i.Public.Ingress.Host,
				}
			}
			if i.Public.LoadBalancer != nil {
				ni.Public.LoadBalancer = &v1alpha2.CapsuleInterfaceLoadBalancer{
					Port: i.Public.LoadBalancer.Port,
				}
			}
		}
		dst.Spec.Interfaces = append(dst.Spec.Interfaces, ni)
	}

	if srcSpec.Replicas != nil {
		dst.Spec.Scale.Horizontal.Instances.Min = uint32(*srcSpec.Replicas)
	}

	if srcSpec.HorizontalScale.MinReplicas != nil {
		dst.Spec.Scale.Horizontal.Instances.Min = *srcSpec.HorizontalScale.MinReplicas
	}

	if srcSpec.HorizontalScale.MaxReplicas != nil {
		dst.Spec.Scale.Horizontal.Instances.Max = ptr.New(*srcSpec.HorizontalScale.MaxReplicas)
	}

	if srcSpec.HorizontalScale.CPUTarget.AverageUtilizationPercentage != 0 {
		dst.Spec.Scale.Horizontal.CPUTarget = &v1alpha2.CPUTarget{
			Utilization: ptr.New(srcSpec.HorizontalScale.CPUTarget.AverageUtilizationPercentage),
		}
	}

	if srcSpec.Resources != nil {
		dst.Spec.Scale.Vertical = &v1alpha2.VerticalScale{
			CPU: &v1alpha2.ResourceLimits{
				Request: srcSpec.Resources.Requests.Cpu(),
				Limit:   srcSpec.Resources.Limits.Cpu(),
			},
			Memory: &v1alpha2.ResourceLimits{
				Request: srcSpec.Resources.Requests.Memory(),
				Limit:   srcSpec.Resources.Limits.Memory(),
			},
			GPU: &v1alpha2.ResourceRequest{
				Request: srcSpec.Resources.Requests["nvidia.com/gpu"],
			},
		}
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1) to this version.
func (dst *Capsule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.Capsule)

	dst.ObjectMeta = src.ObjectMeta

	srcSpec := src.DeepCopy().Spec

	dst.Spec.Args = srcSpec.Args
	dst.Spec.Command = srcSpec.Command
	dst.Spec.Image = srcSpec.Image
	dst.Spec.NodeSelector = srcSpec.NodeSelector

	for _, f := range srcSpec.Files {
		switch {
		case f.Ref != nil:
			switch f.Ref.Kind {
			case "ConfigMap":
				dst.Spec.Files = append(dst.Spec.Files, File{
					ConfigMap: &FileContentRef{
						Name: f.Ref.Name,
						Key:  f.Ref.Key,
					},
					Path: f.Path,
				})
			case "Secret":
				dst.Spec.Files = append(dst.Spec.Files, File{
					Secret: &FileContentRef{
						Name: f.Ref.Name,
						Key:  f.Ref.Key,
					},
					Path: f.Path,
				})
			}
		}
	}

	for _, i := range srcSpec.Interfaces {
		ni := CapsuleInterface{
			Name: i.Name,
			Port: i.Port,
		}
		if i.Public != nil {
			ni.Public = &CapsulePublicInterface{}
			if i.Public.Ingress != nil {
				ni.Public.Ingress = &CapsuleInterfaceIngress{
					Host: i.Public.Ingress.Host,
				}
			}
			if i.Public.LoadBalancer != nil {
				ni.Public.LoadBalancer = &CapsuleInterfaceLoadBalancer{
					Port: i.Public.LoadBalancer.Port,
				}
			}
		}
		dst.Spec.Interfaces = append(dst.Spec.Interfaces, ni)
	}

	dst.Spec.Replicas = ptr.New(int32(srcSpec.Scale.Horizontal.Instances.Min))
	dst.Spec.HorizontalScale.MinReplicas = ptr.New(srcSpec.Scale.Horizontal.Instances.Min)
	dst.Spec.HorizontalScale.MaxReplicas = ptr.New(srcSpec.Scale.Horizontal.Instances.Min)
	if srcSpec.Scale.Horizontal.Instances.Max != nil {
		dst.Spec.HorizontalScale.MaxReplicas = ptr.New(*srcSpec.Scale.Horizontal.Instances.Max)
	}

	if c := srcSpec.Scale.Horizontal.CPUTarget; c != nil && c.Utilization != nil {
		dst.Spec.HorizontalScale.CPUTarget.AverageUtilizationPercentage = *c.Utilization
	}

	if v := srcSpec.Scale.Vertical; v != nil {
		dst.Spec.Resources = &v1.ResourceRequirements{}

		if v.CPU != nil {
			if v.CPU.Request != nil {
				dst.Spec.Resources.Requests[v1.ResourceCPU] = *v.CPU.Request
			}
			if v.CPU.Limit != nil {
				dst.Spec.Resources.Limits[v1.ResourceCPU] = *v.CPU.Limit
			}
		}
		if v.Memory != nil {
			if v.Memory.Request != nil {
				dst.Spec.Resources.Requests[v1.ResourceMemory] = *v.Memory.Request
			}
			if v.Memory.Limit != nil {
				dst.Spec.Resources.Limits[v1.ResourceMemory] = *v.Memory.Limit
			}
		}
		if v.GPU != nil {
			dst.Spec.Resources.Requests["nvidia.com/gpu"] = v.GPU.Request
		}
	}

	return nil
}
