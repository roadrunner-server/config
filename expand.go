package config

import (
	"strings"

	"github.com/spf13/viper"
)

// ExpandVal replaces ${var} or $var in the string based on the mapping function.
// For example, os.ExpandEnv(s) is equivalent to os.Expand(s, os.Getenv).
func ExpandVal(s string, mapping func(string) string) (result string, isChanged bool) {
	var buf []byte
	// ${} is all ASCII, so bytes are fine for this operation.
	isChanged = false
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			if name == "" && w > 0 { //nolint:revive
				// Encountered invalid syntax; eat the
				// characters.
			} else if name == "" {
				// Valid syntax, but $ was not followed by a
				// name. Leave the dollar character untouched.
				buf = append(buf, s[j])
				// parse default syntax
			} else if idx := strings.Index(s, envDefault); idx != -1 {
				// ${key:=default} or ${key:-val}
				substr := strings.Split(name, envDefault)
				if len(substr) != 2 {
					return "", false
				}

				key := substr[0]
				defaultVal := substr[1]

				res := mapping(key)
				if res == "" {
					res = defaultVal
				} else {
					isChanged = true
				}
				buf = append(buf, res...)
			} else {
				buf = append(buf, mapping(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s, isChanged
	}
	return string(buf) + s[i:], isChanged
}

// getShellName returns the name that begins the string and the number of bytes
// consumed to extract it. If the name is enclosed in {}, it's part of a ${}
// expansion and two more bytes are needed than the length of the name.
func getShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		// Scan to closing brace
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2 // Bad syntax; eat "${}"
				}
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; eat "${"
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ { //nolint:revive

	}
	return s[:i], i
}

func expandEnvViper(v *viper.Viper, envFileMap map[string]string) {
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		switch t := val.(type) {
		case string:
			// for string expand it
			v.Set(key, parseEnvDefault(t, envFileMap))
		case []any:
			// for slice -> check if it's a slice of strings
			strArr := make([]string, 0, len(t))
			for i := 0; i < len(t); i++ {
				if valStr, ok := t[i].(string); ok {
					strArr = append(strArr, parseEnvDefault(valStr, envFileMap))
					continue
				}

				v.Set(key, val)
			}

			// we should set the whole array
			if len(strArr) > 0 {
				v.Set(key, strArr)
			}
		default:
			v.Set(key, val)
		}
	}
}

// isShellSpecialVar reports whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore.
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}
