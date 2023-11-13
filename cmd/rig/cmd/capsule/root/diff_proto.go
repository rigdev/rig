package root

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/stuff"
	"github.com/rigdev/rig-go-api/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func printRolloutConfigDiff(root *stuff.KeyNode, nodes []*stuff.RolloutConfigDiffNode) {
	printInner(root, nodes, 0)
}

func printInner(root *stuff.KeyNode, nodes []*stuff.RolloutConfigDiffNode, level int) {
	node := nodes[root.Index]
	indent := strings.Repeat("  ", level)
	if node.Old != nil {
		fmt.Printf("%sOld: %+v\n", indent, node.Old)
	}
	if node.New != nil {
		fmt.Printf("%sNew: %+v\n", indent, node.New)
	}
	for _, node := range root.Children {
		var s string
		switch key := node.Key.Key.(type) {
		case *stuff.Key_Field:
			s = key.Field
		case *stuff.Key_Index:
			s = strconv.Itoa(int(key.Index))
		case *stuff.Key_MapIntKey:
			s = strconv.Itoa(int(key.MapIntKey))
		case *stuff.Key_MapStringKey:
			s = key.MapStringKey
		}
		fmt.Printf("%s%s\n", indent, s)
		printInner(node, nodes, level+1)
	}
}

func rolloutDiff(r1, r2 *capsule.RolloutConfig) *stuff.RolloutConfigDiff {
	parser := newProtoDiffParser(rolloutConfigFieldOneoff, rolloutConfigNodeConstructor)
	key := parser.trav(r1, r2, &stuff.Key{
		Key: nil,
	})
	return &stuff.RolloutConfigDiff{
		Root:  key,
		Nodes: parser.nodes,
	}
}

func rolloutConfigDiffConstructor(root *stuff.KeyNode, nodes []*stuff.RolloutConfigDiffNode) *stuff.RolloutConfigDiff {
	return &stuff.RolloutConfigDiff{
		Root:  root,
		Nodes: nodes,
	}
}

func rolloutConfigNodeConstructor(old, new_ *stuff.RolloutConfig) *stuff.RolloutConfigDiffNode {
	return &stuff.RolloutConfigDiffNode{
		Old: old,
		New: new_,
	}
}

func rolloutConfigFieldOneoff(v any) *stuff.RolloutConfig {
	if v == nil {
		return nil
	}

	res := &stuff.RolloutConfig{
		Field: nil,
	}

	switch v := v.(type) {
	case bool:
		res.Field = &stuff.RolloutConfig_Bool{
			Bool: v,
		}
	case uint32:
		res.Field = &stuff.RolloutConfig_Uint32{
			Uint32: v,
		}
	case uint64:
		res.Field = &stuff.RolloutConfig_Uint64{
			Uint64: v,
		}
	case string:
		res.Field = &stuff.RolloutConfig_String_{
			String_: v,
		}
	case []byte:
		res.Field = &stuff.RolloutConfig_Bytes{
			Bytes: v,
		}
	case *capsule.Authentication:
		res.Field = &stuff.RolloutConfig_Authentication{
			Authentication: v,
		}
	case *capsule.CPUTarget:
		res.Field = &stuff.RolloutConfig_CpuTarget{
			CpuTarget: v,
		}
	case *capsule.Change:
		res.Field = &stuff.RolloutConfig_Change{
			Change: v,
		}
	case *capsule.ConfigFile:
		res.Field = &stuff.RolloutConfig_ConfigFile{
			ConfigFile: v,
		}
	case *capsule.ContainerSettings:
		res.Field = &stuff.RolloutConfig_ContainerSettings{
			ContainerSettings: v,
		}
	case *capsule.GRPC:
		res.Field = &stuff.RolloutConfig_Grpc{
			Grpc: v,
		}
	case *capsule.GRPCMethod:
		res.Field = &stuff.RolloutConfig_GrpcMethod{
			GrpcMethod: v,
		}
	case *capsule.GRPCService:
		res.Field = &stuff.RolloutConfig_GrpcService{
			GrpcService: v,
		}
	case *capsule.GpuLimits:
		res.Field = &stuff.RolloutConfig_GpuLimits{
			GpuLimits: v,
		}
	case *capsule.HorizontalScale:
		res.Field = &stuff.RolloutConfig_HorizontalScale{
			HorizontalScale: v,
		}
	case *capsule.HttpAuth:
		res.Field = &stuff.RolloutConfig_HttpAuth{
			HttpAuth: v,
		}
	case *capsule.Interface:
		res.Field = &stuff.RolloutConfig_Interface{
			Interface: v,
		}
	case *capsule.Logging:
		res.Field = &stuff.RolloutConfig_Logging{
			Logging: v,
		}
	case *capsule.Network:
		res.Field = &stuff.RolloutConfig_Network{
			Network: v,
		}
	case *capsule.PublicInterface:
		res.Field = &stuff.RolloutConfig_PublicInterface{
			PublicInterface: v,
		}
	case *capsule.ResourceList:
		res.Field = &stuff.RolloutConfig_ResourceList{
			ResourceList: v,
		}
	case *capsule.Resources:
		res.Field = &stuff.RolloutConfig_Resources{
			Resources: v,
		}
	case *capsule.RolloutConfig:
		res.Field = &stuff.RolloutConfig_RolloutConfig{
			RolloutConfig: v,
		}
	case *capsule.RoutingMethod:
		res.Field = &stuff.RolloutConfig_RoutingMethod{
			RoutingMethod: v,
		}
	case *timestamppb.Timestamp:
		res.Field = &stuff.RolloutConfig_Timestamp{
			Timestamp: v,
		}
	case *model.Author:
		res.Field = &stuff.RolloutConfig_Author{
			Author: v,
		}

	}
	return res
}

type ProtoDiff[T any] interface {
	GetRoot() *stuff.KeyNode
	GetNodes() []*T
}

type ProtoDiffNode[T any] interface {
	GetOld() *T
	GetNew() *T
}

type protoDiffParser[T any, N ProtoDiffNode[T]] struct {
	constructor     func(v any) *T
	nodeConstructor func(*T, *T) N
	constructor2    func(root *stuff.KeyNode, nodes []N)
	root            *stuff.KeyNode
	nodes           []N
}

func newProtoDiffParser[T any, N ProtoDiffNode[T]](
	constructor func(v any) *T,
	nodeConstructor func(*T, *T) N,
) *protoDiffParser[T, N] {
	return &protoDiffParser[T, N]{
		constructor:     constructor,
		nodeConstructor: nodeConstructor,
		root: &stuff.KeyNode{
			Key:      nil,
			Index:    0,
			Children: []*stuff.KeyNode{},
		},
	}
}

func (p *protoDiffParser[T, N]) trav(v1, v2 any, key *stuff.Key) *stuff.KeyNode {
	nilDiffers := (v1 == nil) != (v2 == nil)
	typeDiffers := reflect.TypeOf(v1) != reflect.TypeOf(v2)
	if nilDiffers || typeDiffers {
		return p.addNodeAndMakeKey(key, v1, v2, nil)
	}

	if v1 == nil {
		return nil
	}

	switch v1 := v1.(type) {
	case bool:
		return p.travPrimitive(v1, v2, key)
	case int32:
		return p.travPrimitive(v1, v2, key)
	case int64:
		return p.travPrimitive(v1, v2, key)
	case uint32:
		return p.travPrimitive(v1, v2, key)
	case uint64:
		return p.travPrimitive(v1, v2, key)
	case float32:
		return p.travPrimitive(v1, v2, key)
	case float64:
		return p.travPrimitive(v1, v2, key)
	case string:
		return p.travPrimitive(v1, v2, key)
	case []byte:
		v2 := v2.([]byte)
		if bytes.Equal(v1, v2) {
			return nil
		}
		return p.addNodeAndMakeKey(key, v1, v2, nil)
	}
	if vv := reflect.ValueOf(v1); vv.Kind() == reflect.Map {
		return p.travMap(v1, v2, key)
	}
	if vv := reflect.ValueOf(v1); vv.Kind() == reflect.Slice {
		return p.travSlice(v1, v2, key)
	}
	return p.travMessage(v1, v2, key)
}

func (p *protoDiffParser[T, N]) travSlice(v1, v2 any, key *stuff.Key) *stuff.KeyNode {
	r1, r2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
	n := r1.Len()
	if r2.Len() < n {
		n = r2.Len()
	}

	var keys []*stuff.KeyNode
	for i := 0; i < n; i++ {
		ff1, ff2 := r1.Index(i).Interface(), r2.Index(i).Interface()
		if keyNode := p.trav(ff1, ff2, &stuff.Key{
			Key: &stuff.Key_Index{
				Index: uint32(i),
			},
		}); keyNode != nil {
			keys = append(keys, keyNode)
		}
	}

	for i := r1.Len(); i < r2.Len(); i++ {
		ff2 := r2.Index(i).Interface()
		if keyNode := p.trav(nil, ff2, &stuff.Key{
			Key: &stuff.Key_Index{
				Index: uint32(i),
			},
		}); keyNode != nil {
			keys = append(keys, keyNode)
		}
	}

	for i := r2.Len(); i < r1.Len(); i++ {
		ff1 := r1.Index(i).Interface()
		if keyNode := p.trav(ff1, nil, &stuff.Key{
			Key: &stuff.Key_Index{
				Index: uint32(i),
			},
		}); keyNode != nil {
			keys = append(keys, keyNode)
		}
	}

	if len(keys) == 0 {
		return nil
	}

	return p.addNodeAndMakeKey(key, nil, nil, keys)
}

func (p *protoDiffParser[T, N]) addNodeAndMakeKey(key *stuff.Key, old, new_ any, childKeys []*stuff.KeyNode) *stuff.KeyNode {
	if old == nil && new_ == nil && len(childKeys) == 0 {
		return nil
	}
	keyNode := &stuff.KeyNode{
		Key:      key,
		Index:    uint32(len(p.nodes)),
		Children: childKeys,
	}
	p.nodes = append(p.nodes, p.nodeConstructor(p.constructor(old), p.constructor(new_)))
	return keyNode
}

func (p *protoDiffParser[T, N]) travMap(v1, v2 any, key *stuff.Key) *stuff.KeyNode {
	r1, r2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
	allKeys := r1.MapKeys()
	for _, k2 := range r2.MapKeys() {
		found := false
		kk2 := k2.Interface()
		for _, k := range r1.MapKeys() {
			kk := k.Interface()
			if kk == kk2 {
				found = true
				break
			}
		}
		if !found {
			allKeys = append(allKeys, k2)
		}
	}

	var childKeys []*stuff.KeyNode
	for _, k := range allKeys {
		var vv1, vv2 any
		if v := r1.MapIndex(k); isNil(v) {
			vv1 = v.Interface()
		}
		if v := r2.MapIndex(k); isNil(v) {
			vv2 = v.Interface()
		}
		if key := p.trav(vv1, vv2, makeMapKey(k.Interface())); key != nil {
			childKeys = append(childKeys, key)
		}
	}

	if len(childKeys) == 0 {
		return nil
	}

	return p.addNodeAndMakeKey(key, nil, nil, childKeys)
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Pointer:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func makeMapKey(key any) *stuff.Key {
	if key, ok := key.(string); ok {
		return &stuff.Key{
			Key: &stuff.Key_MapStringKey{
				MapStringKey: key,
			},
		}
	}

	res := &stuff.Key_MapIntKey{}
	switch key := key.(type) {
	case int8:
		res.MapIntKey = int64(key)
	case int16:
		res.MapIntKey = int64(key)
	case int32:
		res.MapIntKey = int64(key)
	case int64:
		res.MapIntKey = int64(key)
	case int:
		res.MapIntKey = int64(key)
	case uint8:
		res.MapIntKey = int64(key)
	case uint16:
		res.MapIntKey = int64(key)
	case uint32:
		res.MapIntKey = int64(key)
	case uint64:
		res.MapIntKey = int64(key)
	case uint:
		res.MapIntKey = int64(key)
	}

	return &stuff.Key{
		Key: res,
	}
}

func (p *protoDiffParser[T, N]) travMessage(v1, v2 any, key *stuff.Key) *stuff.KeyNode {
	rv1, rv2 := reflect.ValueOf(v1), reflect.ValueOf(v2)
	if isNil(rv1) {
		return nil
	}
	rv1, rv2 = rv1.Elem(), rv2.Elem()
	var childKeys []*stuff.KeyNode
	for _, field := range reflect.VisibleFields(rv1.Type()) {
		if !field.IsExported() {
			continue
		}
		childField1, childField2 := rv1.FieldByIndex(field.Index), rv2.FieldByIndex(field.Index)
		if key := p.trav(childField1.Interface(), childField2.Interface(), &stuff.Key{
			Key: &stuff.Key_Field{
				Field: field.Name,
			},
		}); key != nil {
			childKeys = append(childKeys, key)
		}
	}
	if len(childKeys) == 0 {
		return nil
	}
	return p.addNodeAndMakeKey(key, nil, nil, childKeys)
}

func (p *protoDiffParser[T, N]) travPrimitive(v1, v2 any, key *stuff.Key) *stuff.KeyNode {
	if v1 == v2 {
		return nil
	}
	return p.addNodeAndMakeKey(key, v1, v2, nil)
}
