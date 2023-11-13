package root

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type JSONIndex interface {
	isJSONIndex()
	String() string
}

type JSONKeyIndex struct {
	Value string
}

func (JSONKeyIndex) isJSONIndex() {}

func (j JSONKeyIndex) String() string {
	return j.Value
}

type JSONIntIndex struct {
	Value int
}

func (j JSONIntIndex) String() string {
	return strconv.Itoa(j.Value)
}

func (JSONIntIndex) isJSONIndex() {}

func jsonDiff(v1, v2 any) (JSONDiff, error) {
	bytes1, err := json.Marshal(v1)
	if err != nil {
		return JSONDiff{}, err
	}
	var jsonV1 any
	if err := json.Unmarshal(bytes1, &jsonV1); err != nil {
		return JSONDiff{}, err
	}

	bytes2, err := json.Marshal(v2)
	if err != nil {
		return JSONDiff{}, err
	}
	var jsonV2 any
	if err := json.Unmarshal(bytes2, &jsonV2); err != nil {
		return JSONDiff{}, err
	}

	return jsonDiffInner(jsonV1, jsonV2), nil
}

type JSONDiffChild struct {
	Index JSONIndex
	Diff  JSONDiff
}

type JSONDiff struct {
	Old      any
	New      any
	Children []JSONDiffChild
}

func (j JSONDiff) ToPaths() []JSONDiffPath {
	return j.toPathsInner(nil)
}

func (j JSONDiff) toPathsInner(curPath []JSONIndex) []JSONDiffPath {
	if len(j.Children) == 0 {
		return []JSONDiffPath{{
			Old:  j.Old,
			New:  j.New,
			Path: curPath,
		}}
	}

	var res []JSONDiffPath
	for _, c := range j.Children {
		path := append(curPath[:], c.Index)
		res = append(res, c.Diff.toPathsInner(path)...)
	}

	return res
}

type JSONDiffPath struct {
	Path []JSONIndex
	Old  any
	New  any
}

func (j JSONDiff) Print() {
	j.printInner(0)
}

func (j JSONDiff) printInner(level int) {
	if len(j.Children) != 0 {
		for _, c := range j.Children {
			fmt.Println(strings.Repeat("  ", level) + c.Index.String())
			c.Diff.printInner(level + 1)
		}
		return
	}
	if j.Old != struct{}{} {
		fmt.Println(strings.Repeat("  ", level), "Old:", j.Old)
	}
	if j.New != struct{}{} {
		fmt.Println(strings.Repeat("  ", level), "New:", j.New)
	}
}

func (j JSONDiff) empty() bool {
	return j.Old == nil && j.New == nil && len(j.Children) == 0
}

func jsonDiffInner(v1, v2 any) JSONDiff {
	if (v1 == nil) != (v2 == nil) {
		return JSONDiff{
			Old: v1,
			New: v2,
		}
	}

	switch v1.(type) {
	case bool:
		vv1 := v1.(bool)
		vv2, ok := v2.(bool)
		if !ok || vv1 != vv2 {
			return JSONDiff{
				Old: v1,
				New: v2,
			}
		}
		return JSONDiff{}
	case float64:
		vv1 := v1.(float64)
		vv2, ok := v2.(float64)
		if !ok || vv1 != vv2 {
			return JSONDiff{
				Old: v1,
				New: v2,
			}
		}
		return JSONDiff{}
	case string:
		vv1 := v1.(string)
		vv2, ok := v2.(string)
		if !ok || vv1 != vv2 {
			return JSONDiff{
				Old: v1,
				New: v2,
			}
		}
		return JSONDiff{}
	case []any:
		vv1 := v1.([]any)
		vv2, ok := v2.([]any)
		if !ok {
			return JSONDiff{
				Old: v1,
				New: v2,
			}
		}
		n := len(vv1)
		if len(vv2) < n {
			n = len(vv2)
		}
		var children []JSONDiffChild
		for idx := 0; idx < n; idx++ {
			v, vv := vv1[idx], vv2[idx]
			diff := jsonDiffInner(v, vv)
			if !diff.empty() {
				children = append(children, JSONDiffChild{
					Index: JSONIntIndex{Value: idx},
					Diff:  diff,
				})
			}
		}
		if len(vv1) < len(vv2) {
			for idx, v := range vv2[n:] {
				children = append(children, JSONDiffChild{
					Index: JSONIntIndex{Value: n + idx},
					Diff: JSONDiff{
						Old: struct{}{},
						New: v,
					},
				})
			}
		}
		if len(vv1) > len(vv2) {
			for idx, v := range vv1[n:] {
				children = append(children, JSONDiffChild{
					Index: JSONIntIndex{Value: n + idx},
					Diff: JSONDiff{
						Old: v,
						New: struct{}{},
					},
				})
			}
		}
		return JSONDiff{
			Children: children,
		}
	case map[string]any:
		vv1 := v1.(map[string]any)
		vv2, ok := v2.(map[string]any)
		if !ok {
			return JSONDiff{
				Old: v1,
				New: v2,
			}
		}
		var children []JSONDiffChild
		for k, v := range vv1 {
			vv, ok := vv2[k]
			if !ok {
				children = append(children, JSONDiffChild{
					Index: JSONKeyIndex{Value: k},
					Diff: JSONDiff{
						Old: v,
						New: struct{}{},
					},
				})
				continue
			}
			if diff := jsonDiffInner(v, vv); !diff.empty() {
				children = append(children, JSONDiffChild{
					Index: JSONKeyIndex{Value: k},
					Diff:  diff,
				})
			}
		}
		for k, vv := range vv2 {
			if _, ok := vv1[k]; !ok {
				children = append(children, JSONDiffChild{
					Index: JSONKeyIndex{Value: k},
					Diff: JSONDiff{
						Old: struct{}{},
						New: vv,
					},
				})
			}
		}
		return JSONDiff{
			Children: children,
		}
	}

	return JSONDiff{}
}
