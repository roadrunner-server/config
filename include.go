package config

import (
	"path/filepath"
	"strings"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

func getConfiguration(path, prefix string) (map[string]any, string, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix(prefix)
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := v.ReadInConfig()
	if err != nil {
		return nil, "", err
	}

	// get configuration version
	ver := v.Get(versionKey)
	if ver == nil {
		return nil, "", errors.Str("rr configuration file should contain a version e.g: version: 2.7")
	}

	if _, ok := ver.(string); !ok {
		return nil, "", errors.Errorf("type of version should be string, actual: %T", ver)
	}

	return v.AllSettings(), ver.(string), nil
}

func (p *Plugin) handleInclude(rootVersion string) error {
	ifiles := p.viper.Get(includeKey)
	if ifiles != nil {
		if _, ok := ifiles.([]string); !ok {
			return errors.Str("include should be an array of strings")
		}

		includeFiles := ifiles.([]string)
		for _, file := range includeFiles {
			dir, _ := filepath.Split(p.Path)
			config, version, err := getConfiguration(filepath.Join(dir, file), p.Prefix)
			if err != nil {
				return err
			}

			if version != rootVersion {
				return errors.Str("version in included file must be the same as in root")
			}

			// overriding configuration
			for key, val := range config {
				p.viper.Set(key, val)
			}
		}
	}

	return nil
}
