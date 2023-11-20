package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var rigEnvPrefix = "RIG"
var envReplacer = strings.NewReplacer(".", "_")

type Service interface {
	Get() *v1alpha1.PlatformConfig
}

func NewService(path, secretPath string, scheme *runtime.Scheme) (Service, error) {
	// Get default config.
	cfg := v1alpha1.NewDefaultPlatform()

	bs, err := os.ReadFile(path)
	if err == nil {
		publicCfg, err := deserialize(bs, scheme)
		if err != nil {
			return nil, fmt.Errorf("could not deserialize config: %w", err)
		}

		// overwrite defaults with cfg.
		err = mergo.Merge(cfg, publicCfg, mergo.WithOverride)
		if err != nil {
			return nil, fmt.Errorf("could not merge cfg into default: %w", err)
		}
	}

	var secret *v1alpha1.PlatformConfig
	sbs, err := os.ReadFile(secretPath)
	if err != nil {
		// secrets are in env.
		secret, err = bindEnvSecrets()
		if err != nil {
			return nil, fmt.Errorf("could not bind env secrets: %w", err)
		}
	} else {
		// secrets are in config file.
		secret, err = deserialize(sbs, scheme)
		if err != nil {
			return nil, fmt.Errorf("could not deserialize secret: %w", err)
		}
	}

	// overwrite cfg with secrets.
	err = mergo.Merge(cfg, secret, mergo.WithOverride)
	if err != nil {
		return nil, fmt.Errorf("could not merge secret into config: %w", err)
	}

	return &service{cfg: cfg}, nil
}

type service struct {
	cfg *v1alpha1.PlatformConfig
}

func (s *service) Get() *v1alpha1.PlatformConfig {
	return s.cfg
}

func deserialize(data []byte, scheme *runtime.Scheme) (*v1alpha1.PlatformConfig, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	_, gvk, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("could not decode config: %w", err)
	}

	if gvk.Group != v1alpha1.GroupVersion.Group {
		return nil, errors.New("unsupported api group")
	}

	if gvk.Kind != "PlatformConfig" {
		return nil, errors.New("unsupported api kind")
	}

	cfg := &v1alpha1.PlatformConfig{}

	switch gvk.Version {
	case v1alpha1.GroupVersion.Version:
		if _, _, err := decoder.Decode(data, nil, cfg); err != nil {
			return nil, fmt.Errorf("could not decode into kind: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported api version: %w", err)
	}

	return cfg, nil
}

func bindEnvSecrets() (*v1alpha1.PlatformConfig, error) {
	model := v1alpha1.PlatformConfig{}
	data := make(map[string]interface{})
	constructJSONFromModelAndEnvs(model, &data)

	jsonByte, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal secrets data: %w", err)
	}

	err = json.Unmarshal(jsonByte, &model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// constructs a json format of the model, and then binds the values from envs
func constructJSONFromModelAndEnvs(model interface{}, json *map[string]interface{}, parts ...string) {
	ifv := reflect.ValueOf(model)
	ift := reflect.TypeOf(model)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("json")
		if !ok {
			continue
		}
		tv = strings.TrimSuffix(tv, ",omitempty")
		switch v.Kind() {
		case reflect.Struct:
			// if struct is inline, then append parts to current json level
			if strings.Contains(tv, "inline") {
				constructJSONFromModelAndEnvs(v.Interface(), json)
				continue
			}
			subjson := make(map[string]interface{})
			constructJSONFromModelAndEnvs(v.Interface(), &subjson, append(parts, tv)...)
			// check if subjson interface is empty
			if len(subjson) != 0 {
				(*json)[tv] = subjson
			}
		default:
			p := strings.Join(append(parts, tv), ".")
			stringVal := getEnvValue(p)
			if stringVal != nil {
				(*json)[tv] = stringVal
			}
		}
	}
}

// Fetches the environment value for the given field
func getEnvValue(keyString string) interface{} {
	key := strings.ToUpper(keyString)
	key = envReplacer.Replace(key)
	key = fmt.Sprintf("%s_%s", rigEnvPrefix, key)
	stringVal := os.Getenv(key)
	if stringVal == "" {
		return nil
	}
	// attempt to parse as bool
	boolVal, err := strconv.ParseBool(stringVal)
	if err == nil {
		return boolVal
	}
	// attempt to parse as int
	intVal, err := strconv.ParseInt(stringVal, 10, 64)
	if err == nil {
		return intVal
	}
	// attempt to parse as float
	floatVal, err := strconv.ParseFloat(stringVal, 64)
	if err == nil {
		return floatVal
	}

	// default is string
	return stringVal
}
