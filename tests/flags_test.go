package tests

import (
	"net"
	"testing"
	"time"

	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/stretchr/testify/assert"

	"tests/helpers"
)

// TestFlagOverridesConfigFile checks that a command line override replaces the
// value in the file rather than adding a second listener next to it.
func TestFlagOverridesConfigFile(t *testing.T) {
	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-boot.yaml", []any{&rpcPlugin.Plugin{}, consumer},
		helpers.WithFlags("rpc.listen=tcp://127.0.0.1:6392"),
		helpers.WithTCPProbe("127.0.0.1:6392"),
	)

	assert.Equal(t, "tcp://127.0.0.1:6392", consumer.Snapshot().RPCListen)

	d := net.Dialer{Timeout: time.Second}
	conn, err := d.DialContext(t.Context(), "tcp", "127.0.0.1:6391")
	if err == nil {
		_ = conn.Close()
		assert.Fail(t, "the address from the config file is still bound")
	}
}

// TestFlagValueExpandsEnvDefault checks that an env reference inside an override
// falls back to its default when the variable is not set.
func TestFlagValueExpandsEnvDefault(t *testing.T) {
	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-boot.yaml", []any{&rpcPlugin.Plugin{}, consumer},
		helpers.WithFlags("rpc.listen=tcp://${CONFIG_TEST_RPC_FLAG:-127.0.0.1:6393}"),
		helpers.WithTCPProbe("127.0.0.1:6393"),
	)

	assert.Equal(t, "tcp://127.0.0.1:6393", consumer.Snapshot().RPCListen)
}

func TestBadFlagAbortsContainerInit(t *testing.T) {
	err := helpers.StartExpectInitError(t, "configs/.rr-boot.yaml", []any{&rpcPlugin.Plugin{}, &Consumer{}},
		helpers.WithFlags("rpc.listen="),
	)

	assert.ErrorContains(t, err, "value should not be empty")
}
