package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

func New(filePath string) (Config, error) {
	return new(
		filePath,
		"/etc/rig",
		filepath.Join(os.Getenv("HOME"), ".config", "rig"),
		filepath.Join(os.Getenv("HOME"), ".rig"),
		".",
	)
}

func new(filePath string, searchPaths ...string) (Config, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("RIG")
	v.SetConfigName("server-config")

	if filePath != "" {
		v.SetConfigFile(filePath)
	} else {
		for _, sp := range searchPaths {
			v.AddConfigPath(sp)
		}
	}

	defaultCfg := newDefault()
	bindEnvsAndDefaults(v, defaultCfg)

	if filePath != "" || len(searchPaths) > 0 {
		if err := v.ReadInConfig(); err != nil {
			if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
				return Config{}, fmt.Errorf("could not read in config from viper: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(LogLevelDecodeFunc())); err != nil {
		return Config{}, fmt.Errorf("could not unmarshal loaded viper config: %w", err)
	}

	return cfg, nil
}

func bindEnvsAndDefaults(vi *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindEnvsAndDefaults(vi, v.Interface(), append(parts, tv)...)
		default:
			p := strings.Join(append(parts, tv), ".")
			vi.BindEnv(p)
			if !v.IsZero() {
				vi.SetDefault(p, v.Interface())
			}
		}
	}
}

func LogLevelDecodeFunc() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f == reflect.TypeOf("") && t == reflect.TypeOf(zapcore.DebugLevel) {
			return zapcore.ParseLevel(data.(string))
		}

		return data, nil
	}
}
