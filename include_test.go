package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rootWithIncludes builds a root config in dir that pulls in the given files.
// Include entries are resolved against the working directory, so the paths are
// written out absolute.
func rootWithIncludes(t *testing.T, dir, body string, includes ...string) string {
	t.Helper()

	root := "version: \"3\"\ninclude:\n"
	for _, i := range includes {
		root += fmt.Sprintf("  - %q\n", i)
	}

	return writeFile(t, dir, ".rr.yaml", root+body)
}

func TestIncludeOverridesRoot(t *testing.T) {
	dir := t.TempDir()
	sub := writeFile(t, dir, ".rr-sub.yaml", `version: "3"
logs:
  level: debug
reload:
  interval: 10s
kv:
  roadrunner:
    driver: memory
`)
	root := rootWithIncludes(t, dir, `logs:
  level: info
  mode: development
reload:
  interval: 1s
  patterns: [".php"]
`, sub)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	// The included file wins over the root for the keys it defines.
	assert.Equal(t, "debug", p.Get("logs.level"))
	assert.Equal(t, "10s", p.Get("reload.interval"))

	// Keys only the root defines survive the merge.
	assert.Equal(t, "development", p.Get("logs.mode"))
	assert.Equal(t, []string{".php"}, p.Get("reload.patterns"))

	// A section only the included file defines is reachable, down to a key
	// nested three levels deep.
	assert.True(t, p.Has("kv.roadrunner.driver"))
	assert.Equal(t, "memory", p.Get("kv.roadrunner.driver"))
}

func TestIncludeAppliesFilesInOrder(t *testing.T) {
	dir := t.TempDir()
	first := writeFile(t, dir, ".rr-first.yaml", `version: "3"
logs:
  level: info
`)
	second := writeFile(t, dir, ".rr-second.yaml", `version: "3"
logs:
  level: debug
`)
	root := rootWithIncludes(t, dir, "logs:\n  level: error\n", first, second)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	assert.Equal(t, "debug", p.Get("logs.level"))
}

func TestIncludeUnmarshalKeyPreservesMainOnlySettings(t *testing.T) {
	type poolCfg struct {
		NumWorkers      int    `mapstructure:"num_workers"`
		AllocateTimeout string `mapstructure:"allocate_timeout"`
		DestroyTimeout  string `mapstructure:"destroy_timeout"`
	}

	type pipelineCfg struct {
		Driver string         `mapstructure:"driver"`
		Config map[string]any `mapstructure:"config"`
	}

	type jobsCfg struct {
		Consume      []string               `mapstructure:"consume"`
		PipelineSize int                    `mapstructure:"pipeline_size"`
		Pool         *poolCfg               `mapstructure:"pool"`
		Pipelines    map[string]pipelineCfg `mapstructure:"pipelines"`
	}

	type wholeCfg struct {
		Jobs jobsCfg `mapstructure:"jobs"`
	}

	tests := []struct {
		name             string
		rootBody         string
		includes         []string
		flags            []string
		wantConsume      []string
		wantPipelineSize int
		wantWorkers      int
		wantPipelines    []string
	}{
		{
			name: "reported jobs overlay preserves consume and static pipeline",
			rootBody: `jobs:
  consume: [stand]
  pipelines:
    stand:
      driver: kafka
      config:
        priority: 1
        group_options: { group_id: g }
        consumer_options: { topics: [t] }
`,
			includes: []string{
				`version: "3"
jobs:
  pipeline_size: 100000
  pool: { num_workers: 4, allocate_timeout: 60s, destroy_timeout: 5s }
`,
			},
			wantConsume:      []string{"stand"},
			wantPipelineSize: 100000,
			wantWorkers:      4,
			wantPipelines:    []string{"stand"},
		},
		{
			name: "included value overrides matching key while main-only env expansion survives",
			rootBody: `jobs:
  consume: ["${CONFIG_TEST_MAIN_CONSUME:-stand}"]
  pipeline_size: 10
`,
			includes: []string{
				`version: "3"
jobs:
  pipeline_size: 100000
`,
			},
			wantConsume:      []string{"stand"},
			wantPipelineSize: 100000,
		},
		{
			name: "later include overrides earlier include while keeping earlier-only and main-only keys",
			rootBody: `jobs:
  consume: [stand]
`,
			includes: []string{
				`version: "3"
jobs:
  pipeline_size: 500
  pool: { num_workers: 2 }
`,
				`version: "3"
jobs:
  pipeline_size: 1000
`,
			},
			wantConsume:      []string{"stand"},
			wantPipelineSize: 1000,
			wantWorkers:      2,
		},
		{
			name: "included list replaces main list instead of appending",
			rootBody: `jobs:
  consume: [a, b]
`,
			includes: []string{
				`version: "3"
jobs:
  consume: [c]
  pipeline_size: 200
`,
			},
			wantConsume:      []string{"c"},
			wantPipelineSize: 200,
		},
		{
			name: "included key overrides CLI flag while flag-only key survives",
			rootBody: `jobs:
  consume: [stand]
`,
			includes: []string{
				`version: "3"
jobs:
  pipeline_size: 100
`,
			},
			flags:            []string{"jobs.pipeline_size=50", "jobs.pool.num_workers=2"},
			wantConsume:      []string{"stand"},
			wantPipelineSize: 100,
			wantWorkers:      2,
		},
		{
			name: "single file control",
			rootBody: `jobs:
  consume: [stand]
  pipeline_size: 100000
  pool: { num_workers: 4, allocate_timeout: 60s, destroy_timeout: 5s }
  pipelines:
    stand:
      driver: kafka
      config:
        priority: 1
`,
			wantConsume:      []string{"stand"},
			wantPipelineSize: 100000,
			wantWorkers:      4,
			wantPipelines:    []string{"stand"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			includePaths := make([]string, 0, len(tc.includes))
			for i, body := range tc.includes {
				includePaths = append(includePaths, writeFile(t, dir, fmt.Sprintf(".rr-inc-%d.yaml", i), body))
			}

			var rootPath string
			if len(includePaths) > 0 {
				rootPath = rootWithIncludes(t, dir, tc.rootBody, includePaths...)
			} else {
				rootPath = writeFile(t, dir, ".rr.yaml", "version: \"3\"\n"+tc.rootBody)
			}

			p := &Plugin{Path: rootPath, Flags: tc.flags}
			require.NoError(t, p.Init())

			var section jobsCfg
			require.NoError(t, p.UnmarshalKey("jobs", &section))

			var whole wholeCfg
			require.NoError(t, p.Unmarshal(&whole))

			assert.Equal(t, tc.wantConsume, section.Consume)
			assert.Equal(t, tc.wantConsume, whole.Jobs.Consume)
			assert.Equal(t, tc.wantConsume, p.Get("jobs.consume"))
			assert.Equal(t, tc.wantPipelineSize, section.PipelineSize)
			assert.Equal(t, tc.wantPipelineSize, whole.Jobs.PipelineSize)

			if tc.wantWorkers > 0 {
				require.NotNil(t, section.Pool)
				require.NotNil(t, whole.Jobs.Pool)
				assert.Equal(t, tc.wantWorkers, section.Pool.NumWorkers)
				assert.Equal(t, tc.wantWorkers, whole.Jobs.Pool.NumWorkers)
			}

			if len(tc.wantPipelines) > 0 {
				for _, name := range tc.wantPipelines {
					assert.Contains(t, section.Pipelines, name)
					assert.Contains(t, whole.Jobs.Pipelines, name)
					assert.Equal(t, "kafka", section.Pipelines[name].Driver)
				}
			}
		})
	}
}

func TestIncludeWorksWithoutExperimentalFeatures(t *testing.T) {
	dir := t.TempDir()
	sub := writeFile(t, dir, ".rr-sub.yaml", `version: "3"
rpc:
  listen: tcp://127.0.0.1:6391
`)
	root := rootWithIncludes(t, dir, "logs:\n  level: debug\n", sub)

	p := &Plugin{Path: root, ExperimentalFeatures: false}
	require.NoError(t, p.Init())

	assert.Equal(t, "tcp://127.0.0.1:6391", p.Get("rpc.listen"))
}

func TestIncludeExpandsEnvVars(t *testing.T) {
	t.Setenv("CONFIG_TEST_INCLUDE_LISTEN", "tcp://127.0.0.1:6392")

	dir := t.TempDir()
	sub := writeFile(t, dir, ".rr-sub.yaml", `version: "3"
rpc:
  listen: ${CONFIG_TEST_INCLUDE_LISTEN}
`)
	root := rootWithIncludes(t, dir, "logs:\n  level: debug\n", sub)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	assert.Equal(t, "tcp://127.0.0.1:6392", p.Get("rpc.listen"))
}

func TestIncludeErrors(t *testing.T) {
	tests := []struct {
		name string
		// body is written as the included file; an empty body leaves the path unused.
		body string
		want string
	}{
		{
			name: "missing file",
			body: "",
			want: "no such file",
		},
		{
			name: "no version",
			body: "logs:\n  level: debug\n",
			want: "rr configuration file should contain a version",
		},
		{
			name: "non-string version",
			body: "version: 3\n",
			want: "type of version should be string",
		},
		{
			name: "version differs from the root",
			body: "version: \"2.7\"\nlogs:\n  level: debug\n",
			want: "version in included file must be the same as in root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			sub := filepath.Join(dir, ".rr-sub.yaml")
			if tt.body != "" {
				sub = writeFile(t, dir, ".rr-sub.yaml", tt.body)
			}

			p := &Plugin{Path: rootWithIncludes(t, dir, "", sub)}

			require.ErrorContains(t, p.Init(), tt.want)
		})
	}
}

// The .env tests below reach godotenv, which sets the process environment and
// offers no way back, so every variable name is unique to the test that uses it.

func TestEnvFileLoaded(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env.test", "CONFIG_TEST_ENVFILE_LEVEL=info\nCONFIG_TEST_ENVFILE_INTERVAL=30s\n")
	sub := writeFile(t, dir, ".rr-sub.yaml", `version: "3"
reload:
  interval: ${CONFIG_TEST_ENVFILE_INTERVAL:-1s}
`)
	root := rootWithIncludes(t, dir, `envfile: ".env.test"
logs:
  level: ${CONFIG_TEST_ENVFILE_LEVEL:-error}
`, sub)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	assert.Equal(t, "info", p.Get("logs.level"))

	// The values reach the included files too.
	assert.Equal(t, "30s", p.Get("reload.interval"))
}

// TestEnvFileResolvedRelativeToConfigDir covers the base directory the envfile
// value is joined to: the config file's own directory, which is not the working
// directory the include entries are resolved against.
func TestEnvFileResolvedRelativeToConfigDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "env"), 0o700))
	writeFile(t, filepath.Join(dir, "env"), ".env.test", "CONFIG_TEST_ENVFILE_NESTED_LEVEL=info\n")
	root := writeFile(t, dir, ".rr.yaml", `version: "3"
envfile: "env/.env.test"
logs:
  level: ${CONFIG_TEST_ENVFILE_NESTED_LEVEL:-error}
`)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	assert.Equal(t, "info", p.Get("logs.level"))
}

func TestEnvFileMissingFails(t *testing.T) {
	p := &Plugin{
		Path: writeYAML(t, `version: "3"
envfile: ".env.absent"
`),
	}

	require.ErrorContains(t, p.Init(), "no such file")
}

func TestOSEnvBeatsEnvFile(t *testing.T) {
	t.Setenv("CONFIG_TEST_ENVFILE_PRIORITY_LEVEL", "debug")

	dir := t.TempDir()
	writeFile(t, dir, ".env.test", "CONFIG_TEST_ENVFILE_PRIORITY_LEVEL=info\n")
	root := writeFile(t, dir, ".rr.yaml", `version: "3"
envfile: ".env.test"
logs:
  level: ${CONFIG_TEST_ENVFILE_PRIORITY_LEVEL:-error}
`)

	p := &Plugin{Path: root}
	require.NoError(t, p.Init())

	assert.Equal(t, "debug", p.Get("logs.level"))
}
