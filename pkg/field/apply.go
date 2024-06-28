package field

import (
	"fmt"
	"reflect"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/gonvenience/ytbx"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/proto"
)

func Apply(spec *platformv1.CapsuleSpec, changes []Change) (*platformv1.CapsuleSpec, error) {
	out := proto.Clone(spec).(*platformv1.CapsuleSpec)

	for _, change := range changes {
		path, err := parseJSONPath(change.FieldPath)
		if err != nil {
			return nil, err
		}

		value := reflect.ValueOf(out).Elem()
		var parent reflect.Value

		lookupPath := path.PathElements
		if change.Operation == AddedOperation {
			lookupPath = lookupPath[:len(lookupPath)-1]
		}

		for _, pe := range lookupPath {
			parent = value
			switch value.Type().Kind() {
			case reflect.Slice:
				found := false
				for i := range value.Len() { //nolint:typecheck
					if pe.Key != "" {
						elem := value.Index(i).Elem()
						field, err := findInStruct(elem, pe.Key)
						if err != nil {
							continue
						}

						asStr := fmt.Sprint(field)
						if asStr == pe.Name {
							value = elem
							found = true
							break
						}
					}
				}
				if !found {
					return nil, fmt.Errorf("invalid index in array")
				}

			case reflect.Struct:
				field, err := findInStruct(value, pe.Name)
				if err != nil {
					return nil, err
				}

				if field.Kind() == reflect.Pointer {
					field = field.Elem()
				}
				value = field

			case reflect.Map:
				field := value.MapIndex(reflect.ValueOf(pe.Name))
				if field == (reflect.Value{}) {
					return nil, fmt.Errorf("path '%s' not found", pe.Name)
				}

				// if field.Kind() == reflect.Pointer {
				// 	field = field.Elem()
				// }
				value = field

			default:
				fmt.Println(value.Type().Kind())
				return nil, fmt.Errorf("invalid path navigation: %s", pe.Name)
			}
		}

		switch change.Operation {
		case ModifiedOperation:
			newValue := reflect.New(value.Type())
			if err := obj.Decode([]byte(change.To.AsString), newValue.Interface()); err != nil {
				return nil, err
			}

			if parent.Kind() == reflect.Map {
				parent.SetMapIndex(reflect.ValueOf(lookupPath[len(lookupPath)-1].Name), newValue.Elem())
			} else {
				value.Set(newValue.Elem())
			}

		case RemovedOperation:
			switch parent.Type().Kind() {
			case reflect.Slice:
				for i := range parent.Len() { //nolint:typecheck
					elem := parent.Index(i).Elem()
					if elem == value {
						nval := reflect.MakeSlice(parent.Type(), parent.Len()-1, parent.Len()-1)
						reflect.Copy(nval, parent.Slice(0, i))
						reflect.Copy(nval, parent.Slice(i, parent.Len()))
						parent.Set(nval)
						break
					}
				}

			case reflect.Struct:
			default:
				return nil, fmt.Errorf("invalid asignment")
			}

		case AddedOperation:
			switch value.Type().Kind() {
			case reflect.Slice:
				newValue := reflect.New(value.Type())
				if err := obj.Decode([]byte(change.To.AsString), newValue.Interface()); err != nil {
					return nil, err
				}

				nval := reflect.MakeSlice(value.Type(), value.Len()+1, value.Len()+1)
				reflect.Copy(nval, value)
				nval.Index(value.Len()).Set(newValue.Elem().Index(0))
				value.Set(nval)

			case reflect.Struct:
			default:
				return nil, fmt.Errorf("invalid asignment")
			}
		}
	}

	return out, nil
}

func parseJSONPath(jsonPath string) (ytbx.Path, error) {
	sc := scanner.Scanner{}
	sc.Init(strings.NewReader(jsonPath))
	sc.Error = func(*scanner.Scanner, string) {}
	sc.IsIdentRune = func(r rune, pos int) bool {
		return unicode.IsLetter(r) || r == '_' || (pos > 0 && unicode.IsDigit(r))
	}
	sc.Filename = jsonPath + "\t"

	var result ytbx.Path

	for {
		n := sc.Scan()
		if n == -1 {
			break
		}

		switch sc.TokenText() {
		case ".":
		case "$":
		case "[":
			s, err := parseJSONPathNamed(&sc)
			if err != nil {
				return ytbx.Path{}, err
			}

			result.PathElements = append(result.PathElements, s)
		default:
			result.PathElements = append(result.PathElements, ytbx.PathElement{
				Idx:  -1,
				Name: sc.TokenText(),
			})
		}
	}

	return result, nil
}

func parseJSONPathNamed(sc *scanner.Scanner) (ytbx.PathElement, error) {
	if sc.Scan() != '@' {
		return ytbx.PathElement{}, fmt.Errorf("invalid jsonpath")
	}

	sc.Scan()
	name := sc.TokenText()

	if sc.Scan() != '=' {
		return ytbx.PathElement{}, fmt.Errorf("invalid jsonpath")
	}

	value := ""
	for {
		switch sc.Scan() {
		case 0:
			return ytbx.PathElement{}, fmt.Errorf("invalid jsonpath")
		case ']':
			return ytbx.PathElement{
				Idx:  -1,
				Key:  name,
				Name: value,
			}, nil
		default:
			p := sc.TokenText()
			value += p
		}
	}
}

func findInStruct(value reflect.Value, name string) (reflect.Value, error) {
	for i := range value.Type().NumField() { //nolint:typecheck
		field := value.Type().Field(i)
		if tag, ok := field.Tag.Lookup("json"); ok {
			tagName := strings.Split(tag, ",")[0]
			if tagName == name {
				return value.Field(i), nil
			}
		}
	}

	return reflect.Value{}, fmt.Errorf("property '%s' not found", name)
}
