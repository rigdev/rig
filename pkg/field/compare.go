package field

import (
	"fmt"

	"github.com/gonvenience/ytbx"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type Diff struct {
	From      any
	To        any
	FromBytes []byte
	ToBytes   []byte
	Report    *dyff.Report
	Changes   []Change
}

func Compare(from, to proto.Message, keys ...string) (*Diff, error) {
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

			var paths []nodePath

			var op Operation
			switch det.Kind {
			case dyff.ADDITION:
				op = AddedOperation
				subPaths, err := getNodePaths(det.To, report.To.Documents[0], d.Path, keys)
				if err != nil {
					return nil, err
				}

				paths = subPaths

			case dyff.REMOVAL:
				op = RemovedOperation
				subPaths, err := getNodePaths(det.From, report.From.Documents[0], d.Path, keys)
				if err != nil {
					return nil, err
				}

				paths = subPaths

			case dyff.MODIFICATION:
				op = ModifiedOperation
				paths = []nodePath{{
					path: nil,
				}}
			}

			for _, np := range paths {
				path := append(basePath, np.path...)
				fieldPath := "$"
				fieldID := "$"
				for _, pe := range path {
					if pe.Key != "" && pe.Name != "" {
						fieldPath += fmt.Sprintf("[@%s=%s]", pe.Key, pe.Name)
						fieldID += fmt.Sprintf("[@%s=%s]", pe.Key, pe.Name)
					} else if pe.Name != "" {
						fieldPath += fmt.Sprintf(".%s", pe.Name)
						fieldID += fmt.Sprintf(".%s", pe.Name)
					} else if pe.Idx >= 0 {
						fieldPath += fmt.Sprintf("[%d]", pe.Idx)
						fieldID += fmt.Sprintf("[%d]", pe.Idx)
					} else {
						continue
					}
				}

				change := Change{
					FieldPath: fieldPath,
					FieldID:   fieldID,
					Operation: op,
				}
				if op == ModifiedOperation {
					change.From = GetValue(det.From)
					change.To = GetValue(det.To)
				} else if op == AddedOperation {
					change.To = GetValue(np.node)
				} else if op == RemovedOperation {
					change.From = GetValue(np.node)
				}
				cs = append(cs, change)
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

type nodePath struct {
	path []ytbx.PathElement
	node *yaml.Node
}

func getNodePaths(node, topLevelNode *yaml.Node, currentPath *ytbx.Path, keys []string) ([]nodePath, error) {
	var paths []nodePath

	ytbx.RestructureObject(node)
	switch node.Kind {
	case yaml.SequenceNode:
		for contentIdx, value := range node.Content {
			found := false
			var elemPath ytbx.PathElement
			for _, idKey := range keys {
				altKey, err := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{value},
				}, idKey)
				if err != nil {
					continue
				}
				if altKey != value {
					elemPath = ytbx.PathElement{
						Idx:  -1,
						Key:  idKey,
						Name: altKey.Value,
					}
					found = true
					break
				}
			}
			if !found {
				if len(currentPath.PathElements) == 0 {
					elemPath = ytbx.PathElement{
						Idx: contentIdx,
					}
				} else {
					// TODO: This is a hack to get the index of the value in the sequence.
					// We may need to build on it to get the correct indices throughout the path.
					// It may fail in the case where the sequence is not directly after the base path,
					// and we need to find the element from an index.
					currentNode, err := ytbx.Grab(topLevelNode, currentPath.ToDotStyle())
					if err != nil {
						return nil, err
					}
					for i, c := range currentNode.Content {
						if c == value {
							elemPath = ytbx.PathElement{
								Idx: i,
							}
						}
					}
				}
			}

			paths = append(paths, nodePath{
				path: []ytbx.PathElement{elemPath},
				node: value,
			})
		}
	case yaml.MappingNode:
		for i := range len(node.Content) / 2 { //nolint:typecheck
			key := node.Content[i*2]
			value := node.Content[i*2+1]
			var keyPath ytbx.PathElement
			found := false
			for _, idKey := range keys {
				altKey, err := ytbx.Grab(&yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{value},
				}, idKey)
				if altKey != value && err == nil {
					keyPath = ytbx.PathElement{
						Idx:  -1,
						Key:  idKey,
						Name: altKey.Value,
					}
					found = true
					break
				}
			}
			if !found {
				keyPath = ytbx.PathElement{
					Idx:  -1,
					Name: key.Value,
				}
			}

			subNodePaths, err := getNodePaths(value, topLevelNode, currentPath, keys)
			if err != nil {
				return nil, err
			}
			for _, sub := range subNodePaths {
				paths = append(paths, nodePath{
					path: append([]ytbx.PathElement{keyPath}, sub.path...),
					node: sub.node,
				})
			}
		}

	default:
		paths = append(paths, nodePath{
			path: []ytbx.PathElement{},
			node: node,
		})
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

	r, err := dyff.CompareInputFiles(
		fromFile,
		toFile,
		dyff.KubernetesEntityDetection(false),
		dyff.AdditionalIdentifiers("port", "path", "id", "name"),
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
