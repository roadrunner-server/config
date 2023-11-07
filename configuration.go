package config

import (
	"strings"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/viper"
)

func getConfiguration(path, prefix string) (map[string]any, string, error) {
	const op = errors.Op("sub_config_parsing")
	viper := viper.New()
	viper.AutomaticEnv()
	viper.SetEnvPrefix(prefix)
	viper.SetConfigFile(path)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		return nil, "", errors.E(op, err)
	}

	// get configuration version
	ver := viper.Get(versionKey)
	if ver == nil {
		return nil, "", errors.Str("rr configuration file should contain a version e.g: version: 2.7")
	}

	return viper.AllSettings(), ver.(string), nil
}
