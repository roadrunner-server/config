package tests

import (
	"context"

	"github.com/roadrunner-server/errors"
)

type Foo2 struct {
	configProvider Configurer
}

func (f *Foo2) Init(p Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo2) Serve() chan error {
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

	if allCfg.RPC.Listen != "tcp://127.0.0.1:36643" {
		errCh <- errors.E(op, errors.Str("RPC.Listen should be overwritten"))
		return errCh
	}

	return errCh
}

func (f *Foo2) Stop(context.Context) error {
	return nil
}
