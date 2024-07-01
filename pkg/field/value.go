package field

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Type int

const (
	NullType     Type = iota
	DocumentType Type = iota
	MapType      Type = iota
	ListType     Type = iota
	StringType   Type = iota
	OtherType    Type = iota
)

type Value struct {
	AsString string
	Type     Type
}

func GetValue(node *yaml.Node) Value {
	if node == nil {
		return Value{
			Type: NullType,
		}
	}

	v := Value{
		AsString: node.Value,
	}
	if v.AsString == "" {
		bs, err := yaml.Marshal(node)
		if err != nil {
			panic(err)
		}

		v.AsString = string(bs)
	}

	switch node.Kind {
	case yaml.DocumentNode:
		v.Type = DocumentType

	case yaml.MappingNode:
		v.Type = MapType

	case yaml.SequenceNode:
		v.Type = ListType

	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			v.Type = StringType

		case "!!null":
			v.Type = NullType

		default:
			// use the YAML tag name without the exclamation marks
			v.Type = OtherType // node.Tag[2:]
		}

	case yaml.AliasNode:
		return GetValue(node.Alias)

	default:

		panic(fmt.Errorf("unknown and therefore unsupported kind %v", node.Kind))
	}

	return v
}
