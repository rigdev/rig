package root

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	protoparser "github.com/yoheimuta/go-protoparser"
	"github.com/yoheimuta/go-protoparser/parser"
	"golang.org/x/exp/maps"
)

// Used to generate the proto for the Diff messages

type message struct {
	name string
	pack string
	file string
	m    *parser.Message
}

func (m message) fullName() string {
	if m.pack == "" {
		return m.name
	}
	return m.pack + "." + m.name
}

func parseStuff() error {
	basePath := "/Users/matiasfrank/repos/rig/proto/rig"
	visitor := newVisitor(basePath)
	if err := visitor.run(basePath, "api.v1.capsule.RolloutConfig"); err != nil {
		return err
	}

	return nil
}

var primitiveTypes = []string{
	"bool",
	"uint32",
	"uint64",
	"int32",
	"int64",
	"sint32",
	"sint64",
	"fixed32",
	"fixed64",
	"sfixed32",
	"sfixed64",
	"double",
	"float",
	"string",
	"bytes",
}

type visitor struct {
	basePath       string
	currentFile    string
	currentDir     string
	messages       map[string]message
	visitedTypes   map[string]bool
	currentPackage string
	usedFiles      map[string]bool
	messagePath    []string
}

func newVisitor(basePath string) *visitor {
	return &visitor{
		basePath:     basePath,
		messages:     map[string]message{},
		visitedTypes: map[string]bool{},
		usedFiles:    map[string]bool{},
	}
}

func (v *visitor) run(dir string, baseMessages ...string) error {
	if err := v.readDir(dir); err != nil {
		return err
	}

	for _, m := range baseMessages {
		v.visitMessage(m)
	}

	v.constructProtoString("RolloutConfig")

	return nil
}

func (v *visitor) constructProtoString(name string) string {
	var builder strings.Builder
	builder.WriteString("syntax = \"proto3\";\n\n")

	imports := maps.Keys(v.usedFiles)
	slices.Sort(imports)
	for _, i := range imports {
		builder.WriteString(fmt.Sprintf("import \"%s\";\n", i))
	}
	builder.WriteString("\n")

	builder.WriteString(fmt.Sprintf("message %s {\n", name))
	builder.WriteString(fmt.Sprintf("\toneof field {\n", name))
	types := maps.Keys(v.visitedTypes)
	slices.Sort(types)
	for idx, t := range types {
		split := strings.Split(t, ".")
		tname := split[len(split)-1]
		builder.WriteString(fmt.Sprintf("\t\t%s %s = %v;\n", t, camelToSnake(tname), idx+1))
	}
	builder.WriteString(fmt.Sprintf("\t}\n", name))
	builder.WriteString("}\n")
	return builder.String()
}

// https://stackoverflow.com/questions/56616196/how-to-convert-camel-case-string-to-snake-case
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func (v *visitor) readDir(dir string) error {
	v.basePath = dir
	protoFiles, err := findProtoFiles(dir)
	if err != nil {
		return err
	}

	for _, path := range protoFiles {
		if err := v.visitFile(path); err != nil {
			return err
		}
	}

	return nil
}

func findProtoFiles(dir string) ([]string, error) {
	type entry struct {
		entry fs.DirEntry
		path  string
	}

	var protoFiles []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var queue []entry
	for _, e := range entries {
		queue = append(queue, entry{
			entry: e,
			path:  e.Name(),
		})
	}
	for len(queue) > 0 {
		e := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if e.entry.IsDir() {
			entries, err := os.ReadDir(path.Join(dir, e.path))
			if err != nil {
				return nil, err
			}
			for _, ee := range entries {
				queue = append(queue, entry{
					entry: ee,
					path:  path.Join(e.path, ee.Name()),
				})
			}
		} else {
			if strings.HasSuffix(e.entry.Name(), ".proto") {
				protoFiles = append(protoFiles, e.path)
			}
		}
	}

	return protoFiles, nil
}

func (v *visitor) visitFile(file string) error {
	v.currentPackage = ""
	v.currentFile = file
	v.currentDir = path.Dir(file)
	reader, err := os.Open(path.Join(v.basePath, v.currentFile))
	if err != nil {
		return err
	}
	defer reader.Close()

	got, err := protoparser.Parse(reader)
	if err != nil {
		return err
	}

	for _, el := range got.ProtoBody {
		v.messagePath = nil
		el.Accept(v)
	}

	return nil
}

func (v *visitor) visitMessage(name string) {
	if _, ok := v.visitedTypes[name]; ok {
		return
	}
	v.visitedTypes[name] = true
	message, ok := v.messages[name]
	if !ok {
		return
	}
	v.usedFiles[message.file] = true
	for _, child := range message.m.MessageBody {
		var t string
		switch child := child.(type) {
		case *parser.Field:
			t = child.Type
		case *parser.MapField:
			t = child.Type
		}
		if t == "" {
			continue
		}

		if slices.Contains(primitiveTypes, t) {
			v.visitedTypes[t] = true
			continue
		}

		if !strings.Contains(t, ".") {
			if message.pack != "" {
				t = message.pack + "." + t
			}
		}
		v.visitMessage(t)
	}
}

func (v *visitor) handleType(t string) (string, bool) {
	if slices.Contains(primitiveTypes, t) {
		v.visitedTypes[t] = true
		return "", false
	}

	if !strings.Contains(t, ".") {
		t = strings.Replace(v.currentDir, "/", ".", -1) + "." + t
		return t, true
	}
	return t, false
}

func (*visitor) VisitComment(*parser.Comment)                           {}
func (*visitor) VisitEmptyStatement(*parser.EmptyStatement) (next bool) { return true }
func (*visitor) VisitEnum(*parser.Enum) (next bool)                     { return true }
func (*visitor) VisitEnumField(*parser.EnumField) (next bool)           { return true }
func (*visitor) VisitExtend(*parser.Extend) (next bool)                 { return true }
func (*visitor) VisitExtensions(*parser.Extensions) (next bool)         { return true }
func (v *visitor) VisitField(f *parser.Field) (next bool)               { return true }
func (*visitor) VisitGroupField(*parser.GroupField) (next bool)         { return true }
func (v *visitor) VisitImport(i *parser.Import) (next bool)             { return true }
func (*visitor) VisitMapField(m *parser.MapField) (next bool)           { return true }
func (v *visitor) VisitMessage(m *parser.Message) (next bool) {
	message := message{
		name: m.MessageName,
		pack: v.currentPackage,
		file: v.currentFile,
		m:    m,
	}
	v.messages[message.fullName()] = message
	return true
}
func (*visitor) VisitOneof(*parser.Oneof) (next bool)           { return true }
func (*visitor) VisitOneofField(*parser.OneofField) (next bool) { return true }
func (*visitor) VisitOption(*parser.Option) (next bool)         { return true }
func (v *visitor) VisitPackage(p *parser.Package) (next bool) {
	v.currentPackage = p.Name
	return true
}
func (*visitor) VisitReserved(*parser.Reserved) (next bool) { return true }
func (*visitor) VisitRPC(*parser.RPC) (next bool)           { return true }
func (*visitor) VisitService(*parser.Service) (next bool)   { return true }
func (*visitor) VisitSyntax(*parser.Syntax) (next bool)     { return true }
