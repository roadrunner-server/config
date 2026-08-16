package helpers

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/logger/v6"
	"github.com/stretchr/testify/require"
)

const (
	// ConfigVersion is the RR version the config plugin reports to the plugins
	// that ask for it.
	ConfigVersion = "2023.3.5"
	// probeTimeout caps how long Start waits for the probe to answer.
	probeTimeout = time.Second * 15
	probeTick    = time.Millisecond * 20
	probeDial    = time.Second
)

// bootCfg holds the options applied to a container before it is started.
type bootCfg struct {
	probe func(ctx context.Context) bool
	flags []string
}

// Option customizes the container built by Start and StartExpectInitError.
type Option func(*bootCfg)

// WithFlags passes command line overrides in the `<key>=<value>` form the rr
// binary accepts behind -o.
func WithFlags(flags ...string) Option {
	return func(b *bootCfg) { b.flags = flags }
}

// WithTCPProbe makes Start return only once addr accepts a connection. The rpc
// plugin binds its listener during Serve, so a successful dial proves the
// container reached the address the configuration named.
func WithTCPProbe(addr string) Option {
	return func(b *bootCfg) {
		b.probe = func(ctx context.Context) bool {
			d := net.Dialer{Timeout: probeDial}

			conn, err := d.DialContext(ctx, "tcp", addr)
			if err != nil {
				return false
			}

			return conn.Close() == nil
		}
	}
}

// Start registers the plugins, boots the container and waits for the probe, if
// any, to answer. Errors arriving on the container channel are reported through
// t.Errorf and stop the container, but they do not abort the test. The container
// is stopped by t.Cleanup.
func Start(t *testing.T, cfgPath string, plugins []any, opts ...Option) {
	t.Helper()

	cont, bc := newContainer(t, cfgPath, plugins, opts)
	require.NoError(t, cont.Init())

	ch, err := cont.Serve()
	require.NoError(t, err)

	stopCont := sync.OnceValue(cont.Stop)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}

	wg.Go(func() {
		for {
			select {
			case res := <-ch:
				if res == nil {
					return
				}
				t.Errorf("plugin %s reported an error: %v", res.VertexID, res.Error)
				if errS := stopCont(); errS != nil {
					t.Errorf("container stop: %v", errS)
				}
			case <-done:
				if errS := stopCont(); errS != nil {
					t.Errorf("container stop: %v", errS)
				}

				return
			}
		}
	})

	// The drain goroutine calls t.Errorf, so it has to be joined while the test
	// is still running.
	t.Cleanup(func() {
		close(done)
		wg.Wait()
	})

	if bc.probe != nil {
		require.Eventually(t, func() bool { return bc.probe(t.Context()) }, probeTimeout, probeTick, "container did not become ready")
	}
}

// StartExpectInitError registers the plugins and requires Init to fail, returning its error.
func StartExpectInitError(t *testing.T, cfgPath string, plugins []any, opts ...Option) error {
	t.Helper()

	cont, _ := newContainer(t, cfgPath, plugins, opts)

	err := cont.Init()
	require.Error(t, err)

	return err
}

// newContainer builds the container and registers the config, the logger and the
// caller's plugins. The container is not initialized yet.
func newContainer(t *testing.T, cfgPath string, plugins []any, opts []Option) (*endure.Endure, *bootCfg) {
	t.Helper()

	bc := &bootCfg{}
	for _, o := range opts {
		o(bc)
	}

	all := make([]any, 0, 2+len(plugins))
	all = append(all, &config.Plugin{Version: ConfigVersion, Path: cfgPath, Flags: bc.flags}, &logger.Plugin{})
	all = append(all, plugins...)

	cont := endure.New(slog.LevelDebug)
	require.NoError(t, cont.RegisterAll(all...))

	return cont, bc
}
