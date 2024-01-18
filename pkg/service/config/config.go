package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const envPrefix = "RIG"

type Service interface {
	Operator() *v1alpha1.OperatorConfig
	Platform() *v1alpha1.PlatformConfig
}

func NewService(scheme *runtime.Scheme, filePaths ...string) (Service, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	merger := obj.NewMerger(scheme)

	return newServiceBuilder().
		withDecoder(decoder).
		withFiles(filePaths...).
		withMerger(merger).
		build()
}

func NewServiceFromConfigs(op *v1alpha1.OperatorConfig, platform *v1alpha1.PlatformConfig) Service {
	return &service{
		oCFG: op,
		pCFG: platform,
	}
}

type service struct {
	oCFG *v1alpha1.OperatorConfig
	pCFG *v1alpha1.PlatformConfig
}

func (s *service) Operator() *v1alpha1.OperatorConfig {
	return s.oCFG
}

func (s *service) Platform() *v1alpha1.PlatformConfig {
	return s.pCFG
}

type serviceBuilder struct {
	oCFG      *v1alpha1.OperatorConfig
	pCFG      *v1alpha1.PlatformConfig
	decoder   runtime.Decoder
	merger    obj.Merger
	filePaths []string
}

func newServiceBuilder() *serviceBuilder {
	return &serviceBuilder{
		oCFG: (&v1alpha1.OperatorConfig{}).Default(),
		pCFG: v1alpha1.NewDefaultPlatform(),
	}
}

func (b *serviceBuilder) withDecoder(decoder runtime.Decoder) *serviceBuilder {
	b.decoder = decoder
	return b
}

func (b *serviceBuilder) withFiles(filePaths ...string) *serviceBuilder {
	b.filePaths = append(b.filePaths, filePaths...)
	return b
}

func (b *serviceBuilder) withMerger(merger obj.Merger) *serviceBuilder {
	b.merger = merger
	return b
}

func (b *serviceBuilder) build() (*service, error) {
	for _, filePath := range b.filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not read config file: %w", err)
		}

		if err := b.decode(data); err != nil {
			return nil, err
		}
	}
	var oCFGFromEnv v1alpha1.OperatorConfig
	if err := getCFGFromEnv(&oCFGFromEnv); err != nil {
		return nil, err
	}
	if err := b.merger.Merge(&oCFGFromEnv, b.oCFG); err != nil {
		return nil, fmt.Errorf("could not merge env config: %w", err)
	}

	var pCFGFromEnv v1alpha1.PlatformConfig
	if err := getCFGFromEnv(&pCFGFromEnv); err != nil {
		return nil, err
	}
	if err := b.merger.Merge(&pCFGFromEnv, b.pCFG); err != nil {
		return nil, fmt.Errorf("could not merge env config: %w", err)
	}
	b.oCFG.Default()

	return &service{oCFG: b.oCFG, pCFG: b.pCFG}, nil
}

func (b *serviceBuilder) decode(data []byte) error {
	obj, gvk, err := b.decoder.Decode(data, nil, nil)
	if err != nil {
		return &ErrDecoding{err: err}
	}

	if gvk.Group != v1alpha1.GroupVersion.Group {
		return &ErrUnsupportedGVK{gvk: gvk}
	}

	switch gvk.Kind {
	case "OperatorConfig":
		var decodedCFG *v1alpha1.OperatorConfig
		switch gvk.Version {
		case v1alpha1.GroupVersion.Version:
			cfg, ok := obj.(*v1alpha1.OperatorConfig)
			if !ok {
				return &ErrRuntimeObjectAssertion{
					gvk:    gvk,
					target: "OperatorConfig",
				}
			}
			decodedCFG = cfg
		default:
			return fmt.Errorf("unsupport api version: %s", gvk.Version)
		}
		if err := b.merger.Merge(decodedCFG, b.oCFG); err != nil {
			return fmt.Errorf("could not merge operator config: %w", err)
		}
	case "PlatformConfig":
		var decodedCFG *v1alpha1.PlatformConfig
		switch gvk.Version {
		case v1alpha1.GroupVersion.Version:
			cfg, ok := obj.(*v1alpha1.PlatformConfig)
			if !ok {
				return &ErrRuntimeObjectAssertion{
					gvk:    gvk,
					target: "PlatformConfig",
				}
			}
			decodedCFG = cfg
		default:
			return fmt.Errorf("unsupported api version: %s", gvk.Version)
		}
		if err := b.merger.Merge(decodedCFG, b.pCFG); err != nil {
			return fmt.Errorf("could not merge platform config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported kind: %s", gvk.Kind)
	}

	return nil
}

func getCFGFromEnv(model interface{}) error {
	data := make(map[string]interface{})
	constructJSONFromModelAndEnvs(model, &data)

	jsonByte, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}

	if err := json.Unmarshal(jsonByte, &model); err != nil {
		return fmt.Errorf("could not unmarshal json: %w", err)
	}
	return nil
}

// constructs a json format of the model, and then binds the values from envs
func constructJSONFromModelAndEnvs(model interface{}, json *map[string]interface{}, parts ...string) {
	ifv := reflect.ValueOf(model)
	if ifv.Kind() == reflect.Pointer {
		ifv = ifv.Elem()
	}
	ift := ifv.Type()
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
			p := strings.Join(append(parts, tv), "_")
			val := getEnvValue(p)
			if val != nil {
				(*json)[tv] = val
			}
		}
	}
}

// Fetches the environment value for the given field
func getEnvValue(keyString string) interface{} {
	key := strings.ToUpper(keyString)
	key = fmt.Sprintf("%s_%s", envPrefix, key)
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

	// Parse logging level
	if key == "RIG_LOGGING_LEVEL" {
		level, err := zapcore.ParseLevel(stringVal)
		if err == nil {
			return level
		}

		return zapcore.InfoLevel
	}

	// default is string
	return stringVal
}

type ErrDecoding struct {
	err error
}

func (err *ErrDecoding) Unwrap() error {
	return err.err
}

func (err *ErrDecoding) Error() string {
	return fmt.Sprintf("could not decode config: %s", err.err.Error())
}

type ErrUnsupportedGVK struct {
	gvk *schema.GroupVersionKind
}

func (err *ErrUnsupportedGVK) Error() string {
	return fmt.Sprintf("unsupported group version kind: %s", err.gvk.String())
}

type ErrRuntimeObjectAssertion struct {
	gvk    *schema.GroupVersionKind
	target string
}

func (err *ErrRuntimeObjectAssertion) Error() string {
	return fmt.Sprintf("could not assert %s to %s", err.gvk.String(), err.target)
}
