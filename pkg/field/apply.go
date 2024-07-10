package field

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"

	"github.com/gonvenience/ytbx"
	"github.com/rigdev/rig/pkg/obj"
	"google.golang.org/protobuf/proto"
)

func Apply(object proto.Message, changes []Change) (proto.Message, error) {
	out := proto.Clone(object)

	for _, change := range changes {
		path, err := parseJSONPath(change.FieldPath)
		if err != nil {
			return nil, err
		}

		value := reflect.ValueOf(out).Elem()
		var parent reflect.Value

		lookupPath := path.PathElements

		for _, pe := range lookupPath {
			parent = value
			switch value.Type().Kind() {
			case reflect.Slice:
				// If doing operation on an existing slice where we use the key to find the element
				found := false
				for i := range value.Len() { //nolint:typecheck
					if pe.Key != "" {
						elem := value.Index(i)
						// TODO: This is safe as of now, as we only support proto messages
						if elem.Kind() == reflect.Ptr {
							elem = elem.Elem()
						}
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
				if found {
					continue
				} else if !found && change.Operation == ModifiedOperation || change.Operation == RemovedOperation {
					// If doing a modification or remove operation on existing slice where we find element using the index
					if pe.Idx < 0 || pe.Idx >= value.Len() {
						return nil, fmt.Errorf("invalid index in array")
					}
					value = value.Index(pe.Idx)
					if value.Kind() == reflect.Ptr {
						value = value.Elem()
					}
					continue
				}

				// If adding to a slice, create a reference to the new element
				value = reflect.New(value.Type().Elem()).Elem()

			case reflect.Struct:
				field, err := findInStruct(value, pe.Name)
				if err != nil {
					return nil, err
				}

				// TODO: This is safe as of now, as we only support proto messages - Revisit if it changes
				if field.Kind() == reflect.Pointer {
					if field.IsNil() {
						field.Set(reflect.New(field.Type().Elem()))
					}
					value = field.Elem()
				} else {
					value = field
				}

			case reflect.Map:
				if value.IsNil() {
					value.Set(reflect.MakeMap(value.Type()))
				}

				entry := value.MapIndex(reflect.ValueOf(pe.Name))
				if entry == (reflect.Value{}) && change.Operation != AddedOperation {
					return nil, fmt.Errorf("path '%s' not found", pe.Name)
				}

				value = reflect.New(value.Type().Elem()).Elem()
			default:
				return nil, fmt.Errorf("invalid path navigation: %s", pe.Name)
			}
		}

		switch change.Operation {
		case ModifiedOperation:
			var newValue reflect.Value
			if value.Kind() == reflect.Ptr {
				newValue = reflect.New(value.Type().Elem())
			} else {
				newValue = reflect.New(value.Type())
			}

			if err := obj.Decode([]byte(change.To.AsString), newValue.Interface()); err != nil {
				return nil, err
			}
			switch parent.Type().Kind() {
			case reflect.Map:
				if newValue.Kind() == reflect.Ptr && parent.Type().Elem().Kind() != reflect.Ptr {
					parent.SetMapIndex(reflect.ValueOf(lookupPath[len(lookupPath)-1].Name), newValue.Elem())
				} else {
					parent.SetMapIndex(reflect.ValueOf(lookupPath[len(lookupPath)-1].Name), newValue)
				}
			case reflect.Struct:
				key := path.PathElements[len(path.PathElements)-1].Name
				field, err := findInStruct(parent, key)
				if err != nil {
					return nil, err
				}

				if newValue.Kind() == reflect.Ptr && field.Kind() != reflect.Ptr {
					field.Set(newValue.Elem())
				} else {
					field.Set(newValue)
				}
			case reflect.Slice:
				if newValue.Kind() == reflect.Ptr && value.Kind() != reflect.Ptr {
					value.Set(newValue.Elem())
				} else {
					value.Set(newValue)
				}
			default:
				return nil, fmt.Errorf("invalid assignment")
			}
		case RemovedOperation:
			switch parent.Type().Kind() {
			case reflect.Slice:
				for i := range parent.Len() { //nolint:typecheck
					elem := parent.Index(i).Elem()
					if elem == value {
						nval := reflect.MakeSlice(parent.Type(), parent.Len()-1, parent.Len()-1)
						reflect.Copy(nval, parent.Slice(0, i))
						reflect.Copy(nval, parent.Slice(i+1, parent.Len()))
						parent.Set(nval)
						break
					}
				}

			case reflect.Struct:
				field, err := findInStruct(parent, lookupPath[len(lookupPath)-1].Name)
				if err != nil {
					return nil, err
				}

				field.Set(reflect.New(field.Type()).Elem())
			case reflect.Map:
				parent.SetMapIndex(reflect.ValueOf(lookupPath[len(lookupPath)-1].Name), reflect.Value{})
			default:
				return nil, fmt.Errorf("invalid assignment")
			}

		case AddedOperation:
			if value.Kind() == reflect.Ptr {
				value = reflect.New(value.Type().Elem())
			} else {
				value = reflect.New(value.Type())
			}

			if err := obj.Decode([]byte(change.To.AsString), value.Interface()); err != nil {
				return nil, err
			}

			switch parent.Type().Kind() {
			case reflect.Slice:
				// TODO: Add to the correct index if idx is provided
				nval := reflect.MakeSlice(parent.Type(), parent.Len()+1, parent.Len()+1)
				reflect.Copy(nval, parent)

				if value.Kind() == reflect.Ptr && parent.Type().Elem().Kind() != reflect.Ptr {
					nval.Index(parent.Len()).Set(value.Elem())
				} else {
					nval.Index(parent.Len()).Set(value)
				}

				parent.Set(nval)
			case reflect.Struct:
				key := path.PathElements[len(path.PathElements)-1].Name
				field, err := findInStruct(parent, key)
				if err != nil {
					return nil, err
				}

				if value.Kind() == reflect.Ptr && field.Kind() != reflect.Ptr {
					field.Set(value.Elem())
				} else {
					field.Set(value)
				}
			case reflect.Map:
				key := path.PathElements[len(path.PathElements)-1].Name

				if value.Kind() == reflect.Ptr && parent.Type().Elem().Kind() != reflect.Ptr {
					parent.SetMapIndex(reflect.ValueOf(key), value.Elem())
				} else {
					parent.SetMapIndex(reflect.ValueOf(key), value)
				}
			default:
				return nil, fmt.Errorf("invalid assignment")
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
			s, err := parseJSONPathBracket(&sc)
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

func parseJSONPathBracket(sc *scanner.Scanner) (ytbx.PathElement, error) {
	sc.Scan()
	t := sc.TokenText()
	if t == "@" {
		return parseJSONPathNamed(sc)
	}

	match := _indexRegexp.MatchString(t)
	if match {
		return parseJSONPathIndexed(sc)
	}

	return ytbx.PathElement{}, fmt.Errorf("invalid jsonpath")
}

func parseJSONPathIndexed(sc *scanner.Scanner) (ytbx.PathElement, error) {
	index := sc.TokenText()
	for {
		switch sc.Scan() {
		case ']':
			i, err := strconv.Atoi(index)
			if err != nil {
				return ytbx.PathElement{}, err
			}
			return ytbx.PathElement{
				Idx: i,
			}, nil
		default:
			p := sc.TokenText()
			match := _indexRegexp.MatchString(p)
			if !match {
				return ytbx.PathElement{}, fmt.Errorf("invalid jsonpath")
			}

			index += p
		}
	}
}

func parseJSONPathNamed(sc *scanner.Scanner) (ytbx.PathElement, error) {
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
