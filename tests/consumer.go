package tests

import (
	"context"
	"sync"
)

// Configurer is the subset of the config plugin API a RoadRunner plugin receives
// from the container.
type Configurer interface {
	Get(name string) any
	Has(name string) bool
	Unmarshal(out any) error
	UnmarshalKey(name string, out any) error
	RRVersion() string
}

// Snapshot is what the Configurer answered while the container was building the
// consumer.
type Snapshot struct {
	// RPCListen comes from UnmarshalKey over the rpc section.
	RPCListen string
	// AllRPCListen, LogsMode and LogsLevel come from Unmarshal over the whole
	// configuration.
	AllRPCListen string
	LogsMode     string
	LogsLevel    string
	// RawListen comes from Get, so it carries the value without a struct in
	// between.
	RawListen any
	HasListen bool
	Version   string
}

type rpcSection struct {
	Listen string `mapstructure:"listen"`
}

type wholeConfig struct {
	RPC struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"rpc"`
	Logs struct {
		Mode  string `mapstructure:"mode"`
		Level string `mapstructure:"level"`
	} `mapstructure:"logs"`
}

// Consumer is a plugin that records what the configuration looked like from the
// inside of the container. The expectations live in the tests, never here.
type Consumer struct {
	mu   sync.Mutex
	snap Snapshot
}

func (c *Consumer) Init(cfg Configurer) error {
	var section rpcSection
	if err := cfg.UnmarshalKey("rpc", &section); err != nil {
		return err
	}

	var whole wholeConfig
	if err := cfg.Unmarshal(&whole); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.snap = Snapshot{
		RPCListen:    section.Listen,
		AllRPCListen: whole.RPC.Listen,
		LogsMode:     whole.Logs.Mode,
		LogsLevel:    whole.Logs.Level,
		RawListen:    cfg.Get("rpc.listen"),
		HasListen:    cfg.Has("rpc.listen"),
		Version:      cfg.RRVersion(),
	}

	return nil
}

func (c *Consumer) Serve() chan error {
	return make(chan error, 1)
}

func (c *Consumer) Stop(context.Context) error {
	return nil
}

// Snapshot returns what the Configurer answered during Init.
func (c *Consumer) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.snap
}
