package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandVal(t *testing.T) {
	mapping := func(name string) string {
		return map[string]string{"SET": "val", "PORT": "9000"}[name]
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "bare name", input: "$SET", want: "val"},
		{name: "bare name terminated by punctuation", input: "$SET/x", want: "val/x"},
		{name: "bare name inside text", input: "a $SET b", want: "a val b"},
		{name: "braced name", input: "${SET}", want: "val"},
		{name: "braced name followed by text", input: "${SET}x", want: "valx"},
		{name: "no dollar at all", input: "plain value", want: "plain value"},
		{name: "trailing lone dollar", input: "cost 5$", want: "cost 5$"},
		{name: "dollar followed by a space", input: "$ x", want: "$ x"},
		{name: "empty braces are eaten", input: "${}", want: ""},
		{name: "unterminated brace opener is eaten and the name is left behind", input: "${SET", want: "SET"},
		{name: "special var in braces", input: "${*}", want: ""},
		{name: "special var bare", input: "$1", want: ""},
		{name: "unset name expands to nothing", input: "$MISSING", want: ""},
		{name: "default applies to an unset name", input: "${MISSING:-def}", want: "def"},
		{name: "value wins over the default", input: "${SET:-def}", want: "val"},
		{name: "default inside a larger value", input: "tcp://h:${PORT:-1}", want: "tcp://h:9000"},
		{name: "two defaults in one value", input: "${SCHEME:-tcp}://127.0.0.1:${RPC_PORT:-36643}", want: "tcp://127.0.0.1:36643"},
		{name: "two defaults where one name resolves", input: "${SCHEME:-tcp}://127.0.0.1:${PORT:-36643}", want: "tcp://127.0.0.1:9000"},
		// Only ":-" separates a name from its default, so a bare colon stays part of
		// the name and the lookup finds nothing.
		{name: "colon without a dash is part of the name", input: "${SET:val}", want: ""},
		// A name carrying more than one ":-" splits into three parts and collapses
		// the whole value to an empty string.
		{name: "malformed default", input: "${A:-B:-C}", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExpandVal(tt.input, mapping))
		})
	}
}

// TestExpandValDefaultScope pins the scope of the ":-" check: it is made against
// the whole value, so a single default reference turns every other reference in
// that value into a malformed one and the result collapses to an empty string.
func TestExpandValDefaultScope(t *testing.T) {
	mapping := func(name string) string {
		return map[string]string{"SET": "val"}[name]
	}

	assert.Empty(t, ExpandVal("x${SET}y ${MISSING:-d}", mapping))
}

func TestGetShellName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantWidth int
	}{
		{name: "braced name", input: "{FOO}bar", wantName: "FOO", wantWidth: 5},
		{name: "bare name", input: "FOO bar", wantName: "FOO", wantWidth: 3},
		{name: "special var in braces", input: "{*}", wantName: "*", wantWidth: 3},
		{name: "special var bare", input: "*x", wantName: "*", wantWidth: 1},
		{name: "empty braces", input: "{}", wantName: "", wantWidth: 2},
		{name: "unterminated brace", input: "{FOO", wantName: "", wantWidth: 1},
		{name: "digit is a special var", input: "1abc", wantName: "1", wantWidth: 1},
		{name: "name cannot start with a space", input: " x", wantName: "", wantWidth: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, width := getShellName(tt.input)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantWidth, width)
		})
	}
}

func TestExpandEnvViperTypes(t *testing.T) {
	t.Setenv("CONFIG_EXPAND_HOST", "example.org")

	v := viper.New()
	v.Set("plain", "no references here")
	v.Set("str", "${CONFIG_EXPAND_HOST}")
	v.Set("strs", []any{"${CONFIG_EXPAND_HOST}", "second"})
	v.Set("ints", []any{1, 2})
	v.Set("num", 42)
	v.Set("flag", true)
	v.Set("nested", map[string]any{"host": "${CONFIG_EXPAND_HOST}"})

	expandEnvViper(v)

	assert.Equal(t, "no references here", v.Get("plain"))
	assert.Equal(t, "example.org", v.Get("str"))
	assert.Equal(t, []string{"example.org", "second"}, v.Get("strs"))
	assert.Equal(t, 42, v.Get("num"))
	assert.Equal(t, true, v.Get("flag"))
	assert.Equal(t, "example.org", v.Get("nested.host"))

	// A slice without a single string element keeps its original type.
	assert.Equal(t, []any{1, 2}, v.Get("ints"))
}

// TestExpandEnvViperMixedSlice pins the handling of a slice that mixes strings
// with other types: the string elements are collected into a []string and the
// remaining elements are dropped.
func TestExpandEnvViperMixedSlice(t *testing.T) {
	t.Setenv("CONFIG_EXPAND_MIXED", "expanded")

	v := viper.New()
	v.Set("mixed", []any{"${CONFIG_EXPAND_MIXED}", 7})

	expandEnvViper(v)

	require.Equal(t, []string{"expanded"}, v.Get("mixed"))
}
