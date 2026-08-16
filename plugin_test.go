package config

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rpcConfig is the smallest valid configuration: a version and one section.
const rpcConfig = `version: "3"
rpc:
  listen: tcp://127.0.0.1:6391
`

func TestInitRequiresPath(t *testing.T) {
	p := &Plugin{}

	require.ErrorContains(t, p.Init(), "path should be set")
}

func TestInitMissingFile(t *testing.T) {
	p := &Plugin{Path: filepath.Join(t.TempDir(), ".rr.yaml")}

	require.ErrorContains(t, p.Init(), "no such file")
}

// TestInitFromInlineConfig covers the short circuit taken when the configuration
// arrives as bytes: the file, the flags and the version check are all skipped,
// and the version stays empty instead of falling back to the default.
func TestInitFromInlineConfig(t *testing.T) {
	p := &Plugin{
		Type: "yaml",
		ReadInCfg: []byte(`rpc:
  listen: tcp://127.0.0.1:6391
`),
		Flags: []string{"rpc.listen=tcp://127.0.0.1:6392"},
	}

	require.NoError(t, p.Init())

	assert.Equal(t, "tcp://127.0.0.1:6391", p.Get("rpc.listen"))
	assert.Empty(t, p.RRVersion())
}

func TestInitInlineConfigInvalidYAML(t *testing.T) {
	p := &Plugin{Type: "yaml", ReadInCfg: []byte("rpc: [broken\n")}

	require.ErrorContains(t, p.Init(), "yaml")
}

func TestInitRejectsMissingVersion(t *testing.T) {
	p := &Plugin{Path: writeYAML(t, "rpc:\n  listen: tcp://127.0.0.1:6391\n")}

	require.ErrorContains(t, p.Init(), "rr configuration file should contain a version")
}

func TestInitRejectsNonStringVersion(t *testing.T) {
	p := &Plugin{Path: writeYAML(t, "version: 3\n")}

	err := p.Init()
	require.ErrorContains(t, err, "version should be a string")
	assert.ErrorContains(t, err, "actual type is: int")
}

// TestInitAcceptsPreviousVersion covers the deprecated 2.7 schema: it is still
// accepted, and the schema version is independent of the version the plugin
// reports, which keeps the value it was handed.
func TestInitAcceptsPreviousVersion(t *testing.T) {
	p := &Plugin{
		Path: writeYAML(t, `version: "2.7"
rpc:
  listen: tcp://127.0.0.1:6391
`),
		Version: "2025.1.1",
	}
	require.NoError(t, p.Init())

	assert.Equal(t, "2025.1.1", p.RRVersion())
	assert.Equal(t, "2.7", p.Get("version"))
}

func TestRRVersionDefaulting(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "empty falls back to the default", version: "", want: "3"},
		{name: "local falls back to the default", version: "local", want: "3"},
		{name: "a real version is kept", version: "2025.1.1", want: "2025.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{Path: writeYAML(t, rpcConfig), Version: tt.version}
			require.NoError(t, p.Init())

			assert.Equal(t, tt.want, p.RRVersion())
		})
	}
}

func TestFlagsOverrideFileValues(t *testing.T) {
	t.Run("flag replaces a file value", func(t *testing.T) {
		p := &Plugin{
			Path:  writeYAML(t, rpcConfig),
			Flags: []string{"rpc.listen=tcp://127.0.0.1:6392"},
		}
		require.NoError(t, p.Init())

		assert.Equal(t, "tcp://127.0.0.1:6392", p.Get("rpc.listen"))
	})

	t.Run("flag creates a key the file lacks", func(t *testing.T) {
		p := &Plugin{
			Path:  writeYAML(t, rpcConfig),
			Flags: []string{"logs.level=debug"},
		}
		require.NoError(t, p.Init())

		assert.Equal(t, "debug", p.Get("logs.level"))
	})

	t.Run("flag value expands an env default", func(t *testing.T) {
		p := &Plugin{
			Path:  writeYAML(t, rpcConfig),
			Flags: []string{"rpc.listen=tcp://${CONFIG_TEST_UNSET_HOST:-127.0.0.1:6393}"},
		}
		require.NoError(t, p.Init())

		assert.Equal(t, "tcp://127.0.0.1:6393", p.Get("rpc.listen"))
	})

	// Flags are applied before the version check, so a flag can supply the version
	// a file does not carry.
	t.Run("flag supplies a missing version", func(t *testing.T) {
		p := &Plugin{
			Path:  writeYAML(t, "rpc:\n  listen: tcp://127.0.0.1:6391\n"),
			Flags: []string{"version=3"},
		}
		require.NoError(t, p.Init())

		assert.Equal(t, "3", p.Get("version"))
	})
}

func TestFlagErrorsAbortInit(t *testing.T) {
	tests := []struct {
		name string
		flag string
		want string
	}{
		{name: "no separator", flag: "rpc.listen", want: "invalid flag"},
		{name: "empty key", flag: "=value", want: "key should not be empty"},
		{name: "empty value", flag: "rpc.listen=", want: "value should not be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plugin{Path: writeYAML(t, rpcConfig), Flags: []string{tt.flag}}

			require.ErrorContains(t, p.Init(), tt.want)
		})
	}
}

func TestParseFlag(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		wantKey   string
		wantValue string
	}{
		{name: "plain pair", flag: "a=b", wantKey: "a", wantValue: "b"},
		{name: "surrounding spaces are trimmed", flag: " a = b ", wantKey: "a", wantValue: "b"},
		{name: "value keeps the second separator", flag: "a=b=c", wantKey: "a", wantValue: "b=c"},
		{name: "double quotes are stripped", flag: `a="quoted value"`, wantKey: "a", wantValue: "quoted value"},
		{name: "single quotes are stripped", flag: "a='quoted value'", wantKey: "a", wantValue: "quoted value"},
		{name: "backticks are stripped", flag: "a=`quoted value`", wantKey: "a", wantValue: "quoted value"},
		{name: "escaped quote is unescaped", flag: `a="say \"hi\" now"`, wantKey: "a", wantValue: `say "hi" now`},
		// Init expands the value before storing it, parseFlag hands it over as is.
		{name: "env reference is kept verbatim", flag: "a=${B:-c}", wantKey: "a", wantValue: "${B:-c}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, err := parseFlag(tt.flag)
			require.NoError(t, err)

			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestOverwriteReplacesValues(t *testing.T) {
	p := initFromYAML(t, rpcConfig)

	require.NoError(t, p.Overwrite(map[string]any{
		"rpc.listen": "tcp://127.0.0.1:6392",
		"logs.level": "debug",
	}))

	assert.Equal(t, "tcp://127.0.0.1:6392", p.Get("rpc.listen"))
	assert.Equal(t, "debug", p.Get("logs.level"))
}

func TestPluginAccessors(t *testing.T) {
	p := initFromYAML(t, rpcConfig)

	assert.Equal(t, "config", p.Name())
	assert.True(t, p.Has("rpc.listen"))
	assert.False(t, p.Has("http.address"))
	assert.Nil(t, p.Get("http.address"))
}

func TestGracefulTimeout(t *testing.T) {
	p := &Plugin{Timeout: time.Second * 10}

	assert.Equal(t, time.Second*10, p.GracefulTimeout())
}

func TestUnmarshalKey(t *testing.T) {
	p := initFromYAML(t, rpcConfig)

	out := struct {
		Listen string `mapstructure:"listen"`
	}{}
	require.NoError(t, p.UnmarshalKey("rpc", &out))
	assert.Equal(t, "tcp://127.0.0.1:6391", out.Listen)

	mistyped := struct {
		Listen int `mapstructure:"listen"`
	}{}
	err := p.UnmarshalKey("rpc", &mistyped)
	require.ErrorContains(t, err, "config_plugin_unmarshal_key")
	assert.ErrorContains(t, err, "cannot parse value as 'int'")
}

func TestUnmarshal(t *testing.T) {
	p := initFromYAML(t, rpcConfig)

	out := struct {
		RPC struct {
			Listen string `mapstructure:"listen"`
		} `mapstructure:"rpc"`
	}{}
	require.NoError(t, p.Unmarshal(&out))
	assert.Equal(t, "tcp://127.0.0.1:6391", out.RPC.Listen)

	mistyped := struct {
		RPC struct {
			Listen int `mapstructure:"listen"`
		} `mapstructure:"rpc"`
	}{}
	err := p.Unmarshal(&mistyped)
	require.ErrorContains(t, err, "config_plugin_unmarshal")
	assert.ErrorContains(t, err, "cannot parse value as 'int'")
}

func TestEnvVarsExpandedInFileValues(t *testing.T) {
	t.Setenv("CONFIG_TEST_LISTEN", "tcp://127.0.0.1:6394")

	p := initFromYAML(t, `version: "3"
rpc:
  listen: ${CONFIG_TEST_LISTEN}
logs:
  level: ${CONFIG_TEST_LEVEL:-error}
`)

	assert.Equal(t, "tcp://127.0.0.1:6394", p.Get("rpc.listen"))
	assert.Equal(t, "error", p.Get("logs.level"))
}

func TestEnvVarsExpandedInLists(t *testing.T) {
	t.Setenv("CONFIG_TEST_ADDR_1", "localhost:2999")
	t.Setenv("CONFIG_TEST_ADDR_2", "localhost:2998")

	p := initFromYAML(t, `version: "3"
redis:
  addrs:
    - ${CONFIG_TEST_ADDR_1}
    - ${CONFIG_TEST_ADDR_2}
`)

	assert.Equal(t, []string{"localhost:2999", "localhost:2998"}, p.Get("redis.addrs"))
}
