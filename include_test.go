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
	assert.Equal(t, []any{".php"}, p.Get("reload.patterns"))

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
