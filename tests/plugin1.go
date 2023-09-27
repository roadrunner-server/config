package tests

import (
	"context"
	"time"

	"github.com/roadrunner-server/errors"
)

type Configurer interface {
	GracefulTimeout() time.Duration
	Unmarshal(out any) error
	UnmarshalKey(name string, out any) error
	Has(name string) bool
}

type AllConfig struct {
	RPC struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"rpc"`
	Logs struct {
		Mode  string `mapstructure:"mode"`
		Level string `mapstructure:"level"`
	} `mapstructure:"logs"`
}

type RPCConfig struct {
	Listen string `mapstructure:"listen"`
}

type ServiceConfig struct {
	Enabled   bool
	Recursive bool
	Patterns  []string
	Dirs      []string
	Ignore    []string
}

type Foo struct {
	configProvider Configurer
}

func (f *Foo) Init(p Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo) Serve() chan error {
	const op = errors.Op("foo_plugin_serve")
	errCh := make(chan error, 1)

	r := &RPCConfig{}
	err := f.configProvider.UnmarshalKey("rpc", r)
	if err != nil {
		errCh <- err
	}

	if r.Listen == "" {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	var allCfg AllConfig
	err = f.configProvider.Unmarshal(&allCfg)
	if err != nil {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	if allCfg.RPC.Listen != "tcp://127.0.0.1:6060" {
		errCh <- errors.E(op, errors.Str("RPC.Listen should be parsed"))
		return errCh
	}

	return errCh
}

func (f *Foo) Stop(context.Context) error {
	return nil
}
