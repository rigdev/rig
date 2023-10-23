package uuid

import (
	"encoding/json"
	"reflect"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"gopkg.in/yaml.v3"
)

type UUID string

var Nil UUID = UUID(uuid.Nil.String())

func MustParse(str string) UUID {
	return UUID(uuid.MustParse(str).String())
}

func Parse(str string) (UUID, error) {
	u, err := uuid.Parse(str)
	return UUID(u.String()), err
}

func New() UUID {
	return UUID(uuid.New().String())
}

func (u UUID) IsNil() bool {
	return u == Nil || u == ""
}

func (u UUID) String() string {
	return string(u)
}

// func (u UUID) Bytes() []byte {
// 	return u[:]
// }

func (u UUID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(u.String())
}

func (u *UUID) UnmarshalBSONValue(t bsontype.Type, bs []byte) error {
	v, _, _ := bsoncore.ReadValue(bs, t)
	n, err := Parse(v.StringValue())
	if err != nil {
		return err
	}

	*u = n
	return nil
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *UUID) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	n, err := Parse(str)
	if err != nil {
		return err
	}

	*u = n
	return nil
}

func (u UUID) MarshalYAML() (interface{}, error) {
	if u == Nil {
		return "", nil
	}

	return u.String(), nil
}

func (u *UUID) UnmarshalYAML(v *yaml.Node) error {
	if v.Value == "" {
		*u = Nil
		return nil
	}

	n, err := Parse(v.Value)
	if err != nil {
		return err
	}

	*u = n
	return nil
}

func (u *UUID) Unmarshal(v *yaml.Node) error {
	n, err := Parse(v.Value)
	if err != nil {
		return err
	}

	*u = n
	return nil
}

func MapstructureDecodeFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f == reflect.TypeOf("") && t == reflect.TypeOf(Nil) {
			s := data.(string)

			if s == "" {
				return Nil, nil
			}
			return Parse(s)
		}

		if f == reflect.TypeOf(Nil) && t == reflect.TypeOf("") {
			u := data.(UUID)

			if u == Nil {
				return "", nil
			}

			return u.String(), nil
		}

		return data, nil
	}
}
