package pipeline

import (
	"path"
	"strings"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
)

var _defaultPodAnnotations = []string{RigDevRolloutLabel}

func CreatePodAnnotations(req CapsuleRequest) map[string]string {
	podAnnotations := map[string]string{}
	for _, l := range _defaultPodAnnotations {
		if v, ok := req.Capsule().Annotations[l]; ok {
			podAnnotations[l] = v
		}
	}
	return podAnnotations
}

func FilesToVolumes(files []v1alpha2.File) ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var mounts []v1.VolumeMount
	for _, f := range files {
		volume, mount := FileToVolume(f)
		volumes = append(volumes, volume)
		mounts = append(mounts, mount)
	}
	return volumes, mounts
}

func FileToVolume(f v1alpha2.File) (v1.Volume, v1.VolumeMount) {
	var volume v1.Volume
	var mount v1.VolumeMount
	var name string
	switch f.Ref.Kind {
	case "ConfigMap":
		name = "configmap-" + strings.ReplaceAll(f.Ref.Name, ".", "-")
		volume = v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: f.Ref.Name,
					},
					Items: []v1.KeyToPath{
						{
							Key:  f.Ref.Key,
							Path: path.Base(f.Path),
						},
					},
				},
			},
		}
	case "Secret":
		name = "secret-" + strings.ReplaceAll(f.Ref.Name, ".", "-")
		volume = v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: f.Ref.Name,
					Items: []v1.KeyToPath{
						{
							Key:  f.Ref.Key,
							Path: path.Base(f.Path),
						},
					},
				},
			},
		}
	}
	if name != "" {
		mount = v1.VolumeMount{
			Name:      name,
			MountPath: f.Path,
			SubPath:   path.Base(f.Path),
		}
	}

	return volume, mount
}

func EnvSources(refs []v1alpha2.EnvReference) []v1.EnvFromSource {
	var res []v1.EnvFromSource
	for _, e := range refs {
		switch e.Kind {
		case "ConfigMap":
			res = append(res, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: e.Name},
				},
			})
		case "Secret":
			res = append(res, v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: e.Name},
				},
			})
		}
	}
	return res
}
