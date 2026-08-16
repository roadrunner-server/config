package tests

import (
	"testing"

	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/stretchr/testify/assert"

	"tests/helpers"
)

// TestConfigDrivesRPCListener checks that the address the configuration names is
// both what a plugin reads back and what the rpc plugin actually binds.
func TestConfigDrivesRPCListener(t *testing.T) {
	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-boot.yaml", []any{&rpcPlugin.Plugin{}, consumer}, helpers.WithTCPProbe("127.0.0.1:6391"))

	snap := consumer.Snapshot()
	assert.Equal(t, "tcp://127.0.0.1:6391", snap.RPCListen)
	assert.Equal(t, "tcp://127.0.0.1:6391", snap.AllRPCListen)
	assert.Equal(t, "tcp://127.0.0.1:6391", snap.RawListen)
	assert.True(t, snap.HasListen)
	assert.Equal(t, "development", snap.LogsMode)
	assert.Equal(t, "debug", snap.LogsLevel)
	assert.Equal(t, helpers.ConfigVersion, snap.Version)
}

// TestIncludedConfigReachesPlugins checks that a section defined only in an
// included file reaches the plugins, that an included file wins over the root
// for the keys it repeats, and that the keys only the root defines survive.
func TestIncludedConfigReachesPlugins(t *testing.T) {
	consumer := &Consumer{}
	helpers.Start(t, "configs/.rr-include.yaml", []any{&rpcPlugin.Plugin{}, consumer}, helpers.WithTCPProbe("127.0.0.1:6395"))

	snap := consumer.Snapshot()
	assert.Equal(t, "tcp://127.0.0.1:6395", snap.RPCListen)
	assert.Equal(t, "tcp://127.0.0.1:6395", snap.AllRPCListen)
	assert.Equal(t, "debug", snap.LogsLevel)
	assert.Equal(t, "development", snap.LogsMode)
}
