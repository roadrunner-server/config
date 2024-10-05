package config

import (
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

func getConfiguration(path string) (map[string]any, string, error) {
	v := viper.New()
	v.SetConfigFile(path)
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
	var ifiles []string
	if err := p.viper.UnmarshalKey(includeKey, &ifiles, p.unmarshalOpts()); err != nil {
		return err
	}
	if len(ifiles) == 0 {
		return nil
	}

	for _, file := range ifiles {
		config, version, err := getConfiguration(file)
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
	var envFile string
	if err := p.viper.UnmarshalKey(envFileKey, &envFile, p.unmarshalOpts()); err != nil {
		return err
	}
	if envFile != "" {
		dir, _ := filepath.Split(p.Path)
		return godotenv.Load(filepath.Join(dir, envFile))
	}

	return nil
}
