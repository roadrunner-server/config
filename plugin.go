package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

const (
	PluginName string = "config"
	versionKey string = "version"
	includeKey string = "include"

	defaultConfigVersion string = "3"
	prevConfigVersion    string = "2.7"

	// default envs
	envDefault = ":-"
)

type Plugin struct {
	viper     *viper.Viper
	Path      string
	Prefix    string
	Type      string
	ReadInCfg []byte
	// user defined Flags in the form of <option>.<key> = <value>
	// which overwrites initial a config key
	Flags []string
	// ExperimentalFeatures enables experimental features
	ExperimentalFeatures bool
	// Timeout ...
	Timeout time.Duration
	// RRVersion passed from the Endure.
	Version string
}

// Init config provider.
func (p *Plugin) Init() error {
	const op = errors.Op("config_plugin_init")
	p.viper = viper.New()
	// If user provided []byte data with config, read it and ignore Path and Prefix
	if p.ReadInCfg != nil && p.Type != "" {
		p.viper.SetConfigType("yaml")
		return p.viper.ReadConfig(bytes.NewBuffer(p.ReadInCfg))
	}

	// read in environment variables that match
	p.viper.AutomaticEnv()
	if p.Prefix == "" {
		return errors.E(op, errors.Str("prefix should be set"))
	}

	p.viper.SetEnvPrefix(p.Prefix)
	if p.Path == "" {
		return errors.E(op, errors.Str("path should be set"))
	}

	p.viper.SetConfigFile(p.Path)
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := p.viper.ReadInConfig()
	if err != nil {
		return errors.E(op, err)
	}

	// automatically inject ENV variables using ${ENV} pattern
	expandEnvViper(p.viper)

	// override config Flags
	if len(p.Flags) > 0 {
		for _, f := range p.Flags {
			key, val, errP := parseFlag(f)
			if errP != nil {
				return errors.E(op, errP)
			}
			p.viper.Set(key, parseEnvDefault(val))
		}
	}

	// get a configuration version
	// we should perform this check after all overrides
	ver := p.viper.Get(versionKey)
	if ver == nil {
		return errors.Str("rr configuration file should contain a version e.g: version: 3")
	}

	if _, ok := ver.(string); !ok {
		return errors.E(op, errors.Errorf("version should be a string: `version: \"3\"`, actual type is: %T", ver))
	}

	// hide includes under the experimental flag
	// 'include' is an experimental feature
	// should be here because we need to perform all overrides before
	if p.Experimental() {
		err = p.handleInclude(ver.(string))
		if err != nil {
			return errors.E(op, err)
		}
	}

	// RR includes the config feature by default starting from v2.7.
	// However, this is only required for tests because, starting with v2.7, the rr-binary will pass the version automatically.
	if p.Version == "" || p.Version == "local" {
		p.Version = defaultConfigVersion
	}

	// configuration v2.7
	if ver.(string) == prevConfigVersion {
		println("please, update your configuration version from version: '2.7' to version: '3', see changes here: https://roadrunner.dev/docs/plugins-config/current#v30-configuration")
	}

	return nil
}

// Overwrite overwriting existing config with provided values
func (p *Plugin) Overwrite(values map[string]any) error {
	for key, value := range values {
		p.viper.Set(key, value)
	}

	return nil
}

// Experimental returns true if experimental features are enabled
func (p *Plugin) Experimental() bool {
	return p.ExperimentalFeatures
}

// UnmarshalKey reads a configuration section into a configuration object.
func (p *Plugin) UnmarshalKey(name string, out any) error {
	const op = errors.Op("config_plugin_unmarshal_key")
	err := p.viper.UnmarshalKey(name, &out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (p *Plugin) Unmarshal(out any) error {
	const op = errors.Op("config_plugin_unmarshal")
	err := p.viper.Unmarshal(&out)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// Get raw config in the form of a config section.
func (p *Plugin) Get(name string) any {
	return p.viper.Get(name)
}

// Has checks if a config section exists.
func (p *Plugin) Has(name string) bool {
	return p.viper.IsSet(name)
}

// RRVersion returns current RR version
func (p *Plugin) RRVersion() string {
	return p.Version
}

func (p *Plugin) GracefulTimeout() time.Duration {
	return p.Timeout
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func parseFlag(flag string) (string, string, error) {
	const op = errors.Op("parse_flag")
	if !strings.Contains(flag, "=") {
		return "", "", errors.E(op, errors.Errorf("invalid flag `%s`", flag))
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)
	if len(parts) < 2 {
		return "", "", errors.Str("usage: -o key=value")
	}

	if parts[0] == "" {
		return "", "", errors.Str("key should not be empty")
	}

	if parts[1] == "" {
		return "", "", errors.Str("value should not be empty")
	}

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.ReplaceAll(value, fmt.Sprintf("\\%s", string(escape)), string(escape))
	}

	return value
}

func parseEnvDefault(val string) string {
	// tcp://127.0.0.1:${RPC_PORT:-36643}
	// for envs like this, part would be tcp://127.0.0.1:
	return ExpandVal(val, os.Getenv)
}
