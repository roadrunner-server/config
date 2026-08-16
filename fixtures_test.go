package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeFile writes body into dir under name and returns the full path.
func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	return path
}

// writeYAML writes body as a config file in a fresh temp directory and returns
// its path. Paths are absolute, so include lists composed from them resolve the
// same way regardless of the working directory.
func writeYAML(t *testing.T, body string) string {
	t.Helper()

	return writeFile(t, t.TempDir(), ".rr.yaml", body)
}

// initFromYAML initializes a plugin over body and requires Init to succeed.
func initFromYAML(t *testing.T, body string) *Plugin {
	t.Helper()

	p := &Plugin{Path: writeYAML(t, body)}
	require.NoError(t, p.Init())

	return p
}
