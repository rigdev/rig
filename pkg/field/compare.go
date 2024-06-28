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
			basePath := d.Path.PathElements

			var paths [][]ytbx.PathElement

			var op Operation
			switch det.Kind {
			case dyff.ADDITION:
				op = AddedOperation
				subPaths, err := getNodePaths(det.To)
				if err != nil {
					return nil, err
				}

				paths = subPaths

			case dyff.REMOVAL:
				op = RemovedOperation
				subPaths, err := getNodePaths(det.From)
				if err != nil {
					return nil, err
				}

				paths = subPaths

			case dyff.MODIFICATION:
				op = ModifiedOperation
				paths = [][]ytbx.PathElement{nil}
			}

			for _, p := range paths {
				path := append(basePath, p...)
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

func getNodePaths(node *yaml.Node) ([][]ytbx.PathElement, error) {
	var paths [][]ytbx.PathElement

	ytbx.RestructureObject(node)
	switch node.Kind {
	case yaml.SequenceNode:
		for i, value := range node.Content {
			found := false
			for _, idKey := range []string{"port", "path"} {
				altKey, _ := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{value},
				}, idKey)
				if altKey != value {
					paths = append(paths, []ytbx.PathElement{{
						Idx:  i,
						Key:  idKey,
						Name: altKey.Value,
					}})
					found = true
					break
				}
			}
			if !found {
				paths = append(paths, []ytbx.PathElement{{
					Idx: i,
				}})
			}
		}

	case yaml.MappingNode:
		singleElement := len(node.Content) == 2
		for i := range len(node.Content) / 2 {
			key := node.Content[i*2]
			value := node.Content[i*2+1]
			found := false
			var keyPath ytbx.PathElement
			for _, idKey := range []string{"port", "path"} {
				altKey, err := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{value},
				}, idKey)
				if altKey != value && err == nil {
					keyPath = ytbx.PathElement{
						Idx:  i,
						Key:  idKey,
						Name: altKey.Value,
					}
					found = true
					break
				}
			}
			if !found {
				keyPath = ytbx.PathElement{
					Idx:  i,
					Name: key.Value,
				}
			}

			if singleElement {
				subNodePaths, err := getNodePaths(value)
				if err != nil {
					return nil, err
				}
				for _, sub := range subNodePaths {
					paths = append(paths, append([]ytbx.PathElement{keyPath}, sub...))
				}
			} else {
				paths = append(paths, []ytbx.PathElement{keyPath})
			}
		}

	default:
		paths = append(paths, []ytbx.PathElement{{}})
	}

	return paths, nil
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
