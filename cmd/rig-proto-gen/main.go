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

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	v1 "github.com/rigdev/rig/pkg/api/platform/v1"
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
		{pkg: "encoding/json", name: "RawMessage"}: {
			message: &protoMessage{
				file: "google/protobuf/struct.proto",
				pkg:  "google/protobuf",
				name: "Struct",
			},
		},
	}
	existingProtoFile = map[string]struct{}{
		"google/protobuf/struct.proto": struct{}{},
	}
	packageRewrites = map[string]string{
		"github.com/rigdev/rig/pkg/api/v1alpha2":        "v1alpha2",
		"github.com/rigdev/rig/pkg/api/platform/v1":     "platform/v1",
		"github.com/rigdev/rig/pkg/api/config/v1alpha1": "config/v1alpha1",
	}
	compiledTypes = []any{
		v1.CapsuleSet{},
		v1.Capsule{},
		v1.Environment{},
		v1.Project{},
		v1.HostCapsule{},
		v1alpha1.OperatorConfig{},
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
		var b strings.Builder
		for _, file := range pkg.files {
			if _, ok := existingProtoFile[file.name]; ok || pkg.name == "" {
				continue
			}
			file.compile(&b)
			p := path.Join(outputDir, file.name)
			fmt.Println("writing", p, "name", file.name)
			if err := os.MkdirAll(path.Dir(p), 0o777); err != nil {
				return err
			}
			if err := os.WriteFile(p, []byte(strings.TrimSuffix(b.String(), "\n")), 0o777); err != nil {
				return err
			}
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
		if _, _, err := p.parseType(reflect.TypeOf(v)); err != nil {
			return err
		}
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

func (p *parser) parseType(t reflect.Type) (protoType, []protoType, error) {
	pt, pts, err := p.parseTypeInner(t)
	if err != nil {
		return protoType{}, nil, err
	}
	if err := pt.validate(); err != nil {
		return protoType{}, nil, err
	}
	return pt, pts, nil
}

func (p *parser) parseTypeInner(t reflect.Type) (protoType, []protoType, error) {
	pkgName := rewritePackage(t.PkgPath())
	if known, ok := knownMessages[messageType{
		pkg:  pkgName,
		name: t.Name(),
	}]; ok {
		if known.builtIn != nil {
			pt := protoType{
				builtIn: known.builtIn,
			}
			return pt, nil, nil
		} else if known.message != nil {
			p.getPackage(known.message.pkg).getFile(known.message.file).getMessage(known.message.name)
			return known.message.geType(), []protoType{known.message.geType()}, nil
		}
	}

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
		return protoType{}, nil, fmt.Errorf("can't pass type %s", t.Name())
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		pt, pts, err := p.parseType(t.Elem())
		if err != nil {
			return protoType{}, nil, err
		}
		if pt.mapType != nil {
			return protoType{}, nil, fmt.Errorf("cannot have slice map")
		} else if pt.repeated != nil {
			return protoType{}, nil, fmt.Errorf("cannot have slice of slice")
		}
		res := protoType{
			repeated: &repeatedType{
				builtIn: pt.builtIn,
				message: pt.message,
			},
		}
		return res, pts, nil
	case reflect.Pointer:
		return p.parseType(t.Elem())
	case reflect.Map:
		keyType, keyDeps, err := p.parseType(t.Key())
		if err != nil {
			return protoType{}, nil, err
		}
		if keyType.builtIn == nil {
			return protoType{}, nil, fmt.Errorf("map key value must be integral or string")
		}

		valueType, valueDeps, err := p.parseType(t.Elem())
		if err != nil {
			return protoType{}, nil, err
		}
		if valueType.mapType != nil {
			return protoType{}, nil, fmt.Errorf("map value cannot be a map")
		}
		if valueType.repeated != nil {
			return protoType{}, nil, fmt.Errorf("map value cannot be repeated")
		}

		return protoType{
			mapType: &mapType{
				key: *keyType.builtIn,
				value: mapValueType{
					builtIn: valueType.builtIn,
					message: valueType.message,
				},
			},
		}, append(keyDeps, valueDeps...), nil
	case reflect.Struct:
		pkgName := rewritePackage(t.PkgPath())
		if p.hasMessage(pkgName, t.Name()) {
			msg := p.getPackage(pkgName).getMessage(t.Name())
			return msg.geType(), []protoType{msg.geType()}, nil
		}
		pkg := p.getPackage(pkgName)
		file := pkg.getFile(path.Join(pkgName, GENERATED_FILE_NAME))
		msg := file.getMessage(t.Name())
		for _, field := range extractFields(t) {
			protoIdx, err := getProtoIndex(field)
			if err != nil {
				return protoType{}, nil, fmt.Errorf("field %s on type %s had no wellformed proto index tag: %s", field.Name, t.Name(), err)
			}
			pt, deps, err := p.parseType(field.Type)
			if err != nil {
				return protoType{}, nil, err
			}
			for _, tt := range deps {
				if tt.message != nil {
					file.addImport(tt.message.file)
				}
			}
			msg.fields = append(msg.fields, protoField{
				t:    pt,
				name: jsonName(field),
				idx:  protoIdx,
			})
		}
		return msg.geType(), []protoType{msg.geType()}, nil
	default:
		return protoType{
			builtIn: &builtintType{t: t.Kind().String()},
		}, nil, nil
	}
}

func (p *parser) getPackage(pkg string) *protoPackage {
	pp, ok := p.packages[pkg]
	if !ok {
		pp = &protoPackage{
			name:  pkg,
			files: map[string]*protoFile{},
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
	name  string
	files map[string]*protoFile
}

type protoFile struct {
	pkg          string
	name         string
	imports      map[string]struct{}
	messages     map[string]*protoMessage
	messageOrder []string
}

func (p *protoFile) getMessage(name string) *protoMessage {
	m, ok := p.messages[name]
	if !ok {
		m = &protoMessage{
			pkg:  p.pkg,
			file: p.name,
			name: name,
		}
		p.messages[name] = m
		p.messageOrder = append(p.messageOrder, name)
	}
	return m
}

func (p *protoPackage) validate() error {
	allMessages := map[string]struct{}{}
	for _, f := range p.files {
		for _, m := range f.messages {
			if _, ok := allMessages[m.name]; ok {
				return fmt.Errorf("message %s defined multiple times", m.name)
			}
			if err := m.validate(); err != nil {
				return fmt.Errorf("message %s has errors: %q", m.name, err)
			}
			allMessages[m.name] = struct{}{}
		}
	}
	return nil
}

func (p *protoFile) compile(b *strings.Builder) {
	b.WriteString("syntax = \"proto3\";\n")
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("package %s;\n", formatPkg(p.pkg)))
	b.WriteString("\n")

	imports := maps.Keys(p.imports)
	slices.Sort(imports)
	for _, imp := range imports {
		b.WriteString(fmt.Sprintf("import \"%s\";\n", imp))
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

func (p *protoPackage) getFile(name string) *protoFile {
	f, ok := p.files[name]
	if !ok {
		f = &protoFile{
			pkg:      p.name,
			name:     name,
			imports:  map[string]struct{}{},
			messages: map[string]*protoMessage{},
		}
		p.files[name] = f
	}
	return f
}

func (p *protoPackage) getMessage(name string) *protoMessage {
	for _, f := range p.files {
		if f.hasMessage(name) {
			return f.getMessage(name)
		}
	}
	return nil
}

func (p *protoPackage) hasMessage(name string) bool {
	for _, f := range p.files {
		if f.hasMessage(name) {
			return true
		}
	}
	return false
}

func (p *protoFile) addImport(path string) {
	if path != "" && path != p.name {
		p.imports[path] = struct{}{}
	}
}

func (p *protoFile) hasMessage(name string) bool {
	_, ok := p.messages[name]
	return ok
}

type protoMessage struct {
	pkg    string
	file   string
	name   string
	fields []protoField
}

func (p protoMessage) geType() protoType {
	return protoType{
		message: &messageType{
			pkg:  p.pkg,
			file: p.file,
			name: p.name,
		},
	}
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
	file string
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
	builtIn *builtintType
	message *messageType
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
	if m.builtIn != nil {
		m.builtIn.compile(b)
	} else if m.message != nil {
		m.message.compile(b, curPkg)
	}
}

type repeatedType struct {
	builtIn *builtintType
	message *messageType
}

func (r repeatedType) validate() error {
	c := 0
	if r.builtIn != nil {
		c++
	}
	if r.message != nil {
		c++
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
	}
}
