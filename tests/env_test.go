package tests

import (
	"testing"

	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/stretchr/testify/assert"

	"tests/helpers"
)

// TestEnvVarSetsListenAddress checks that an env reference in the config file
// reaches the rpc plugin as the address it binds.
func TestEnvVarSetsListenAddress(t *testing.T) {
	t.Setenv("CONFIG_TEST_RPC_ADDR", "tcp://127.0.0.1:6394")

	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-env.yaml", []any{&rpcPlugin.Plugin{}, consumer}, helpers.WithTCPProbe("127.0.0.1:6394"))

	assert.Equal(t, "tcp://127.0.0.1:6394", consumer.Snapshot().RPCListen)
}

// TestEnvDefaultAppliesWhenUnset checks the fallback written into the config
// file for variables the environment does not carry.
func TestEnvDefaultAppliesWhenUnset(t *testing.T) {
	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-env-default.yaml", []any{&rpcPlugin.Plugin{}, consumer}, helpers.WithTCPProbe("127.0.0.1:6396"))

	snap := consumer.Snapshot()
	assert.Equal(t, "tcp://127.0.0.1:6396", snap.RPCListen)
	assert.Equal(t, "development", snap.LogsMode)
}
