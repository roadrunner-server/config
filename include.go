package config

import (
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

func getConfiguration(path, prefix string, envFileMap map[string]string) (map[string]any, string, error) {
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

	// automatically inject ENV variables using ${ENV} pattern
	expandEnvViper(v, envFileMap)

	return v.AllSettings(), ver.(string), nil
}

func (p *Plugin) handleInclude(rootVersion string) error {
	ifiles := p.viper.GetStringSlice(includeKey)
	if ifiles == nil {
		return nil
	}

	for _, file := range ifiles {
		dir, _ := filepath.Split(p.Path)
		config, version, err := getConfiguration(filepath.Join(dir, file), p.Prefix, p.envFileMap)
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

	return nil
}

func (p *Plugin) handleEnvFile() error {
	envFile := p.viper.GetString(envFileKey)
	if envFile != "" {
		var err error
		dir, _ := filepath.Split(p.Path)
		p.envFileMap, err = godotenv.Read(filepath.Join(dir, envFile))
		if err != nil {
			return err
		}
	}

	return nil
}
