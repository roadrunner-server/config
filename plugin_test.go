package config

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvArr(t *testing.T) {
	require.NoError(t, syscall.Setenv("REDIS_HOST_1", "localhost:2999"))
	require.NoError(t, syscall.Setenv("REDIS_HOST_2", "localhost:2998"))
	p := &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-env-arr.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.NoError(t, err)

	str := p.viper.Get("redis.addrs")
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
	// rr 2.8, config 2.7
	p := &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.yaml",
		Version: "2.8",
	}

	err := p.Init()
	require.NoError(t, err)

	// rr 2.7, config 2.8
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.8.yaml",
		Version: "2.7",
	}

	err = p.Init()
	require.Error(t, err)

	// rr 2.7, config no version
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-no-version.yaml",
		Version: "2.7",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.8, config 2.8
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.8.yaml",
		Version: "2.8",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.8, config 2.8
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.8.yaml",
		Version: "2.8",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.7, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.7",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.7, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.7",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.8.1, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.8.1",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.5, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.5",
	}

	err = p.Init()
	require.Error(t, err)

	// rr 2.7.3, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.7.3",
	}

	err = p.Init()
	require.NoError(t, err)

	// no version but overwrite
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-no-version.yaml",
		Version: "2.8.1",
	}
	p.Flags = append(p.Flags, "version=2.7")

	err = p.Init()
	require.NoError(t, err)

	// rr 2.10.0, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.10.0",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.10.3, config 2.7.3
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.3.yaml",
		Version: "2.10.3",
	}

	err = p.Init()
	require.NoError(t, err)

	// no version but overwrite
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-no-version.yaml",
		Version: "2.10.1",
	}
	p.Flags = append(p.Flags, "version=2.7")

	err = p.Init()
	require.NoError(t, err)

	// rr 2.10, config 2.7
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.7.yaml",
		Version: "2.10",
	}

	err = p.Init()
	require.NoError(t, err)

	// rr 2.10, config 2.8
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-2.11.yaml",
		Version: "2.10",
	}

	err = p.Init()
	require.Error(t, err)

	// rr 2.10, config no version
	p = &Plugin{
		Prefix:  "rr",
		Path:    "tests/.rr-no-version.yaml",
		Version: "2.10",
	}

	err = p.Init()
	require.NoError(t, err)
}
