package tests

import (
	"syscall"
	"testing"

	"github.com/roadrunner-server/config/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvArr(t *testing.T) {
	require.NoError(t, syscall.Setenv("REDIS_HOST_1", "localhost:2999"))
	require.NoError(t, syscall.Setenv("REDIS_HOST_2", "localhost:2998"))
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr-env-arr.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.NoError(t, err)

	str := p.Get("redis.addrs")
	if _, ok := str.([]string); !ok {
		t.Fatal("not a slice")
	}

	require.Len(t, str.([]string), 2)

	if str.([]string)[0] != "localhost:2999" && str.([]string)[0] != "localhost:2998" {
		t.Fatalf("not expanded")
	}

	if str.([]string)[1] != "localhost:2999" && str.([]string)[1] != "localhost:2998" {
		t.Fatalf("not expanded")
	}
}

func TestVersions(t *testing.T) {
	// rr 2.10, config no version
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr-no-version.yaml",
		Version: "2.10",
	}

	err := p.Init()
	require.Error(t, err)
}

func TestIncludingConfigs(t *testing.T) {
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr.include1.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.NoError(t, err)

	// Values from .rr.include1.yaml (root file)
	assert.Equal(t, []string([]string{".php"}), p.Get("reload.patterns"))

	// Values from .rr.include1-sub1.yaml (1st included file)
	assert.Equal(t, "127.0.0.1:15389", p.Get("http.address"))
	assert.Equal(t, "10s", p.Get("reload.interval")) // Checking if overrided

	// Values from .rr.include1-sub2.yaml (2nd included file)
	assert.Equal(t, "console", p.Get("logs.encoding"))
}

func TestErrorWhenIncludedConfigHaveDifferentVersionThenRoot(t *testing.T) {
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr.include2.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config_plugin_init: version in included file must be the same like in root")
}
