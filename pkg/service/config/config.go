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
	"github.com/spf13/afero"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
)

const envPrefix = "RIG"

func NewPlatformConfig(fs afero.Fs, scheme *runtime.Scheme, options ...Option) (*v1alpha1.PlatformConfig, error) {
	builder := newConfigBuilder(fs, scheme, options...)
	config, err := build(builder, v1alpha1.NewDefaultPlatform(), &v1alpha1.PlatformConfig{})
	if err != nil {
		return nil, err
	}
	config.Migrate()
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func NewOperatorConfig(fs afero.Fs, scheme *runtime.Scheme, options ...Option) (*v1alpha1.OperatorConfig, error) {
	builder := newConfigBuilder(fs, scheme, options...)
	config, err := build(builder, (&v1alpha1.OperatorConfig{}).Default(), &v1alpha1.OperatorConfig{})
	if err != nil {
		return nil, err
	}
	return config.Default(), nil
}

func newConfigBuilder(fs afero.Fs, scheme *runtime.Scheme, options ...Option) *configBuilder {
	c := &configBuilder{
		scheme:     scheme,
		fs:         fs,
		serializer: obj.NewSerializer(scheme),
	}
	for _, o := range options {
		o(c)
	}

	return c
}

type Option func(c *configBuilder)

func WithFilePaths(paths ...string) Option {
	return func(c *configBuilder) {
		c.filePaths = paths
	}
}

func WithContent(content string) Option {
	return func(c *configBuilder) {
		c.content = content
	}
}

type configBuilder struct {
	scheme     *runtime.Scheme
	fs         afero.Fs
	serializer runtime.Serializer
	filePaths  []string
	content    string
}

func build[T runtime.Object](c *configBuilder, defaults, emptyInit T) (T, error) {
	var empty T
	var err error
	result := defaults
	for _, filePath := range c.filePaths {
		data, err := afero.ReadFile(c.fs, filePath)
		if err != nil {
			return empty, err
		}

		result, err = merge(data, result, c.scheme, c.serializer)
		if err != nil {
			return empty, err
		}
	}

	if c.content != "" {
		result, err = merge([]byte(c.content), result, c.scheme, c.serializer)
		if err != nil {
			return empty, err
		}
	}

	envPatch := emptyInit.DeepCopyObject()
	if err := getCFGFromEnv(envPatch); err != nil {
		return empty, err
	}

	return obj.Merge(envPatch, result, result, c.serializer)
}

func merge[T runtime.Object](data []byte, object T, scheme *runtime.Scheme, serializer runtime.Serializer) (T, error) {
	var empty T

	// TODO Enforce no unknown fields!

	patch, err := obj.DecodeAnyRuntime(data, scheme)
	if err != nil {
		return empty, err
	}
	res, err := obj.Merge(patch, object, object, serializer)
	return res, err
}

func getCFGFromEnv(model any) error {
	data := make(map[string]any)
	constructJSONFromModelAndEnvs(model, &data)

	jsonByte, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonByte, &model); err != nil {
		return err
	}
	return nil
}

// constructs a json format of the model, and then binds the values from envs
func constructJSONFromModelAndEnvs(model any, json *map[string]any, parts ...string) {
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
			subjson := make(map[string]any)
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
func getEnvValue(keyString string) any {
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
