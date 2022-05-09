package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
