package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"

	v1 "github.com/rigdev/rig/pkg/api/platform/v1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type knownType struct {
	builtIn *builtintType
	message *protoMessage
}

var (
	GENERATED_FILE_NAME = "generated.proto"
	knownMessages       = map[messageType]knownType{
		{pkg: "k8s.io/apimachinery/pkg/api/resource", name: "Quantity"}: {
			builtIn: &builtintType{
				t: "string",
			},
		},
	}
	packageRewrites = map[string]string{
		"github.com/rigdev/rig/pkg/api/v1alpha2":    "v1alpha2",
		"github.com/rigdev/rig/pkg/api/platform/v1": "platform/v1",
	}
	compiledTypes = []any{
		v1.CapsuleSet{},
		v1.Capsule{},
		v1.Environment{},
		v1.Project{},
		v1.HostCapsule{},
		v1alpha2.CapsuleScale{},
	}
)

func main() {
	cmd := &cobra.Command{
		Use:           "proto-gen outputdir",
		Short:         "Tool for generating proto files from Kubernetes CRDs",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          run,
	}
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(_ *cobra.Command, args []string) error {
	outputDir := args[0]
	p := &parser{
		packages: map[string]*protoPackage{},
	}

	if err := p.parse(compiledTypes...); err != nil {
		return err
	}

	for _, pkg := range p.packages {
		if err := pkg.validate(); err != nil {
			return fmt.Errorf("package %s had errors: %s", pkg.name, err)
		}
	}

	for _, pkg := range p.packages {
		if pkg.name == "" {
			continue
		}
		var b strings.Builder
		pkg.compile(&b)
		filePath := path.Join(outputDir, pkg.name, GENERATED_FILE_NAME)
		fmt.Println("writing", filePath)
		if err := os.MkdirAll(path.Dir(filePath), 0o777); err != nil {
			return err
		}
		if err := os.WriteFile(filePath, []byte(b.String()), 0o777); err != nil {
			return err
		}
	}

	return nil
}

func formatPkg(pkg string) string {
	pkg = strings.Replace(pkg, "-", "_", -1)
	pkg = strings.Replace(pkg, "/", ".", -1)
	return pkg
}

func rewritePackage(pkg string) string {
	if name, ok := packageRewrites[pkg]; ok {
		return name
	}
	return pkg
}

type parser struct {
	packages map[string]*protoPackage
}

func (p *parser) parse(values ...any) error {
	for _, v := range values {
		if err := p.parseMessage(reflect.TypeOf(v)); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) parseMessage(t reflect.Type) error {
	if t.Kind() != reflect.Struct {
		return nil
	}

	pkgName := rewritePackage(t.PkgPath())
	if known, ok := knownMessages[messageType{
		pkg:  pkgName,
		name: t.Name(),
	}]; ok {
		if known.builtIn != nil {
			return nil
		} else if known.message != nil {
			pkg := p.getPackage(pkgName)
			msg := pkg.getMessage(t.Name())
			msg.fields = known.message.fields
		}

		return nil
	}

	pkg := p.getPackage(pkgName)
	msg := pkg.getMessage(t.Name())

	for _, field := range extractFields(t) {
		protoIdx, err := getProtoIndex(field)
		if err != nil {
			return fmt.Errorf("field %s on type %s had no wellformed proto index tag: %s", field.Name, t.Name(), err)
		}
		types, protoType, err := unravelType(field.Type)
		if err != nil {
			return err
		}
		for _, tt := range types {
			pkgName := rewritePackage(tt.PkgPath())
			if !p.hasMessage(pkgName, tt.Name()) {
				if err := p.parseMessage(tt); err != nil {
					return err
				}
			}
			pkg.addImport(pkgName)
		}
		msg.fields = append(msg.fields, protoField{
			t:    protoType,
			name: jsonName(field),
			idx:  protoIdx,
		})
	}

	return nil
}

func getProtoIndex(f reflect.StructField) (int, error) {
	tag, ok := f.Tag.Lookup("protobuf")
	if !ok {
		return 0, errors.New("no protobuf tag")
	}
	splits := strings.Split(tag, ",")
	var ints []int
	for _, s := range splits {
		if i, err := strconv.Atoi(s); err == nil {
			ints = append(ints, i)
		}
	}
	if len(ints) != 1 {
		return 0, fmt.Errorf("expected exactly one integer element of the protobuf tag, got %v", len(ints))
	}

	return ints[0], nil
}

func extractFields(t reflect.Type) []reflect.StructField {
	var res []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if isJSONInline(jsonTag) {
			tt := field.Type
			for tt.Kind() == reflect.Pointer {
				tt = tt.Elem()
			}
			if tt.Kind() != reflect.Struct {
				continue
			}
			res = append(res, extractFields(field.Type)...)
		} else {
			res = append(res, field)
		}
	}

	return res
}

func isJSONInline(s string) bool {
	// TODO More robust
	return s == ",inline"
}

func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func unravelType(t reflect.Type) ([]reflect.Type, protoType, error) {
	var repeated bool
	var message *messageType
	var builtIn *builtintType
	var mapT *mapType
	var types []reflect.Type
	for {
		done := false
		switch t.Kind() {
		case reflect.Chan:
			fallthrough
		case reflect.Func:
			fallthrough
		case reflect.Interface:
			fallthrough
		case reflect.Invalid:
			fallthrough
		case reflect.UnsafePointer:
			return nil, protoType{}, fmt.Errorf("can't pass type %s", t.Name())
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			if repeated {
				return nil, protoType{}, fmt.Errorf("can't pass nested slice")
			}
			repeated = true
			t = t.Elem()
		case reflect.Pointer:
			t = t.Elem()
		case reflect.Map:
			keyT, keyType, err := unravelType(t.Key())
			if err != nil {
				return nil, protoType{}, err
			}
			if keyType.builtIn == nil {
				return nil, protoType{}, fmt.Errorf("map key value must be integral or string")
			}
			valueT, valueType, err := unravelType(t.Elem())
			if err != nil {
				return nil, protoType{}, err
			}
			types = append(types, keyT...)
			types = append(types, valueT...)
			var mapVT mapValueType
			if valueType.builtIn != nil {
				mapVT = mapValueType{builtIn: valueType.builtIn}
			} else if valueType.message != nil {
				mapVT = mapValueType{message: valueType.message}
			} else if valueType.repeated != nil {
				if valueType.repeated.mapType != nil {
					return nil, protoType{}, fmt.Errorf("map value cannot be a map")
				}
				mapVT = mapValueType{
					builtIn:  valueType.repeated.builtIn,
					message:  valueType.repeated.message,
					repeated: true,
				}
			} else if valueType.mapType != nil {
				return nil, protoType{}, fmt.Errorf("map value cannot be a map")
			}
			mapT = &mapType{
				key:   *keyType.builtIn,
				value: mapVT,
			}
			done = true
		case reflect.Struct:
			msg := messageType{
				pkg:  rewritePackage(t.PkgPath()),
				name: t.Name(),
			}
			if known, ok := knownMessages[msg]; ok {
				if known.builtIn != nil {
					builtIn = known.builtIn
				} else if known.message != nil {
					message = &messageType{
						pkg:  known.message.pkg,
						name: known.message.name,
					}
				}
				done = true
			} else {
				message = &msg
				types = append(types, t)
			}
			done = true
		default:
			builtIn = &builtintType{t: t.Kind().String()}
			types = append(types, t)
			done = true
		}
		if done {
			break
		}
	}

	var result protoType
	if repeated {
		result = protoType{
			repeated: &repeatedType{
				builtIn: builtIn,
				message: message,
				mapType: mapT,
			},
		}
	} else {
		result = protoType{
			builtIn: builtIn,
			message: message,
			mapType: mapT,
		}
	}

	if err := result.validate(); err != nil {
		return nil, protoType{}, err
	}

	return types, result, nil
}

func (p *parser) getPackage(pkg string) *protoPackage {
	pp, ok := p.packages[pkg]
	if !ok {
		pp = &protoPackage{
			name:     pkg,
			imports:  map[string]struct{}{},
			messages: map[string]*protoMessage{},
		}
		p.packages[pkg] = pp
	}
	return pp
}

func (p *parser) hasMessage(pkg string, name string) bool {
	return p.getPackage(pkg).hasMessage(name)
}

func (p *parser) getMessage(pkg string, name string) *protoMessage {
	return p.getPackage(pkg).getMessage(name)
}

type protoPackage struct {
	name         string
	imports      map[string]struct{}
	messages     map[string]*protoMessage
	messageOrder []string
}

func (p *protoPackage) validate() error {
	for _, m := range p.messages {
		if err := m.validate(); err != nil {
			return fmt.Errorf("message %s has errors: %q", m.name, err)
		}
	}

	return nil
}

func (p *protoPackage) compile(b *strings.Builder) {
	b.WriteString("syntax = \"proto3\";\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("package %s;\n", formatPkg(p.name)))
	b.WriteString("\n")

	imports := maps.Keys(p.imports)
	slices.Sort(imports)
	for _, imp := range imports {
		b.WriteString(fmt.Sprintf("import \"%s/%s\";\n", imp, GENERATED_FILE_NAME))
	}
	if len(p.imports) > 0 {
		b.WriteString("\n")
	}

	for _, name := range p.messageOrder {
		msg := p.messages[name]
		msg.compile(b)
		b.WriteString("\n")
	}
}

func (p *protoPackage) getMessage(name string) *protoMessage {
	m, ok := p.messages[name]
	if !ok {
		m = &protoMessage{
			pkg:  p.name,
			name: name,
		}
		p.messages[name] = m
		p.messageOrder = append(p.messageOrder, name)
	}
	return m
}

func (p *protoPackage) hasMessage(name string) bool {
	_, ok := p.messages[name]
	return ok
}

func (p *protoPackage) addImport(path string) {
	if path != "" && path != p.name {
		p.imports[path] = struct{}{}
	}
}

type protoMessage struct {
	pkg    string
	name   string
	fields []protoField
}

func (p protoMessage) validate() error {
	indexes := map[int]struct{}{}
	for _, field := range p.fields {
		if _, ok := indexes[field.idx]; ok {
			return fmt.Errorf("multiple fields with the same proto index %v", field.idx)
		}
		if field.idx < 1 {
			return fmt.Errorf("invalid proto index %v", field.idx)
		}
		indexes[field.idx] = struct{}{}
	}
	return nil
}

func (p protoMessage) compile(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("message %s {\n", p.name))
	for _, field := range p.fields {
		field.compile(b, p.pkg)
	}
	b.WriteString("}\n")
}

type protoField struct {
	t    protoType
	name string
	idx  int
}

func (p protoField) compile(b *strings.Builder, curPkg string) {
	b.WriteString("  ")
	p.t.compile(b, curPkg)
	b.WriteString(fmt.Sprintf(" %s = %v;\n", p.name, p.idx))
}

type protoType struct {
	builtIn  *builtintType
	message  *messageType
	repeated *repeatedType
	mapType  *mapType
}

func (p protoType) compile(b *strings.Builder, curPkg string) {
	if p.builtIn != nil {
		p.builtIn.compile(b)
		return
	} else if p.message != nil {
		p.message.compile(b, curPkg)
	} else if p.repeated != nil {
		p.repeated.compile(b, curPkg)
	} else if p.mapType != nil {
		p.mapType.compile(b, curPkg)
	}
}

func (p protoType) validate() error {
	c := 0
	if p.builtIn != nil {
		c++
	}
	if p.message != nil {
		c++
	}
	if p.repeated != nil {
		c++
		if err := p.repeated.validate(); err != nil {
			return err
		}
	}
	if p.mapType != nil {
		c++
		if err := p.mapType.validate(); err != nil {
			return err
		}
	}

	if c != 1 {
		return fmt.Errorf("protoType must have exactly one field set")
	}
	return nil
}

type builtintType struct {
	t string
}

func (b builtintType) compile(bb *strings.Builder) {
	var s string
	switch b.t {
	case "float32":
		s = "float"
	case "float64":
		s = "double"
	case "uint8":
		fallthrough
	case "uint16":
		s = "uint32"
	case "int8":
		fallthrough
	case "int16":
		s = "int32"
	case "uint":
		s = "uint64"
	case "int":
		s = "int"
	default:
		s = b.t
	}
	bb.WriteString(s)
}

type messageType struct {
	pkg  string
	name string
}

func (m messageType) compile(b *strings.Builder, curPkg string) {
	if curPkg == m.pkg {
		b.WriteString(m.name)
	} else {
		b.WriteString(fmt.Sprintf("%s.%s", formatPkg(m.pkg), m.name))
	}
}

type mapType struct {
	key   builtintType
	value mapValueType
}

func (m mapType) validate() error {
	return m.value.validate()
}

func (m mapType) compile(b *strings.Builder, curPkg string) {
	b.WriteString("map<")
	m.key.compile(b)
	b.WriteString(", ")
	m.value.compile(b, curPkg)
	b.WriteString(">")
}

type mapValueType struct {
	builtIn  *builtintType
	message  *messageType
	repeated bool
}

func (m mapValueType) validate() error {
	c := 0
	if m.builtIn != nil {
		c++
	}
	if m.message != nil {
		c++
	}

	if c != 1 {
		return errors.New("mapValueType must have exactly ony of builtIn and message set")
	}

	return nil
}

func (m mapValueType) compile(b *strings.Builder, curPkg string) {
	if m.repeated {
		b.WriteString("repeated ")
	}
	if m.builtIn != nil {
		m.builtIn.compile(b)
	} else if m.message != nil {
		m.message.compile(b, curPkg)
	}
}

type repeatedType struct {
	builtIn *builtintType
	message *messageType
	mapType *mapType
}

func (r repeatedType) validate() error {
	c := 0
	if r.builtIn != nil {
		c++
	}
	if r.message != nil {
		c++
	}
	if r.mapType != nil {
		c++
		if err := r.mapType.validate(); err != nil {
			return err
		}
	}

	if c != 1 {
		return errors.New("repeatedType must have exactly one field set.")
	}

	return nil
}

func (r repeatedType) compile(b *strings.Builder, curPkg string) {
	if r.builtIn != nil {
		if r.builtIn.t == "uint8" {
			b.WriteString("bytes")
		} else {
			b.WriteString("repeated ")
			r.builtIn.compile(b)
		}
	} else if r.message != nil {
		b.WriteString("repeated ")
		r.message.compile(b, curPkg)
	} else if r.mapType != nil {
		b.WriteString("repeated ")
		r.mapType.compile(b, curPkg)
	}
}
