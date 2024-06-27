package field

import (
	"fmt"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/pkg/obj"
	"gopkg.in/yaml.v3"
)

type Diff struct {
	From      *platformv1.CapsuleSpec
	To        *platformv1.CapsuleSpec
	FromBytes []byte
	ToBytes   []byte
	Report    *dyff.Report
	Changes   []Change
}

func Compare(from *platformv1.CapsuleSpec, to *platformv1.CapsuleSpec) (*Diff, error) {
	fromBytes, err := obj.EncodeAny(from)
	if err != nil {
		return nil, err
	}

	toBytes, err := obj.EncodeAny(to)
	if err != nil {
		return nil, err
	}

	report, err := compareBytes(fromBytes, toBytes)
	if err != nil {
		return nil, err
	}

	var cs []Change

	for _, d := range report.Diffs {
		for _, det := range d.Details {
			path := d.Path.PathElements

			var op Operation
			switch det.Kind {
			case dyff.ADDITION:
				op = AddedOperation
				subPath, err := getNodePath(det.To)
				if err != nil {
					return nil, err
				}

				path = append(path, subPath...)

			case dyff.REMOVAL:
				op = RemovedOperation
				subPath, err := getNodePath(det.From)
				if err != nil {
					return nil, err
				}

				path = append(path, subPath...)
			case dyff.MODIFICATION:
				op = ModifiedOperation
			}

			fieldPath := "$"
			fieldID := "$"
			for _, p := range path {
				if p.Key != "" {
					fieldPath += fmt.Sprintf("[@%s=%s]", p.Key, p.Name)
				} else {
					fieldPath += "." + p.Name
				}

				if p.Key != "" {
					fieldID += fmt.Sprintf("[@%s=]", p.Key)
				} else {
					fieldID += "." + p.Name
				}
			}

			cs = append(cs, Change{
				FieldPath: fieldPath,
				FieldID:   fieldID,
				Operation: op,
				From:      GetValue(det.From),
				To:        GetValue(det.To),
			})
		}
	}

	return &Diff{
		From:      from,
		To:        to,
		FromBytes: fromBytes,
		ToBytes:   toBytes,
		Report:    report,
		Changes:   cs,
	}, nil
}

func getNodePath(node *yaml.Node) ([]ytbx.PathElement, error) {
	var path []ytbx.PathElement
	subPath, err := ytbx.ListStringKeys(node)
	if err != nil {
		return nil, err
	}
	for _, sp := range subPath {
		path = append(path, ytbx.PathElement{
			Idx:  -1,
			Name: sp,
		})
	}

	ytbx.RestructureObject(node)
	switch node.Kind {
	case yaml.SequenceNode:
		for i, item := range node.Content[0].Content {
			for _, idKey := range []string{"port", "path"} {
				key, err := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{item},
				}, idKey)
				if err == nil {
					path = append(path, ytbx.PathElement{
						Idx:  i,
						Key:  idKey,
						Name: key.Value,
					})
					break
				}
			}
		}

	case yaml.MappingNode:
		for i, item := range node.Content[1].Content {
			for _, idKey := range []string{"port", "path"} {
				key, err := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{item},
				}, idKey)
				if err == nil {
					path = append(path, ytbx.PathElement{
						Idx:  i,
						Key:  idKey,
						Name: key.Value,
					})
					break
				}
			}
		}
	}

	return path, nil
}

func compareBytes(from, to []byte) (*dyff.Report, error) {
	if len(from) == 0 {
		from = []byte("{}")
	}
	if len(to) == 0 {
		to = []byte("{}")
	}

	fromNodes, err := ytbx.LoadYAMLDocuments(from)
	if err != nil {
		return nil, err
	}
	fromFile := ytbx.InputFile{
		Location:  "from",
		Documents: fromNodes,
	}
	toNodes, err := ytbx.LoadYAMLDocuments(to)
	if err != nil {
		return nil, err
	}
	toFile := ytbx.InputFile{
		Location:  "to",
		Documents: toNodes,
	}

	r, err := dyff.CompareInputFiles(fromFile, toFile, dyff.KubernetesEntityDetection(false), dyff.AdditionalIdentifiers("port", "path"))
	if err != nil {
		return nil, err
	}

	return &r, nil
}
