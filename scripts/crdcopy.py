s = "\n  versions:\n"

with open("deploy/charts/rig-operator/templates/crd.yaml") as f:
    helm_data = f.read()
helm_idx = helm_data.find(s) + len(s)
helm_data = helm_data[:helm_idx]

with open("deploy/kustomize/crd/bases/rig.dev_capsules.yaml") as f:
    crd_data = f.read()
crd_idx = crd_data.find(s) + len(s)
crd_data = crd_data[crd_idx:]

result = helm_data + crd_data + "{{- end }}\n"
with open("deploy/charts/rig-operator/templates/crd.yaml", "w") as f:
    f.write(result)
