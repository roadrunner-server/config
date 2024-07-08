package tests

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	configImpl "github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/logger/v5"
	"github.com/roadrunner-server/rpc/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cfgPath = "configs/.rr.yaml"

func TestViperProvider_Init(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = nil

	err := cont.Register(vp)
	require.NoError(t, err)

	err = cont.Register(&Foo{})
	require.NoError(t, err)

	err = cont.Init()
	require.NoError(t, err)

	ch, err := cont.Serve()
	require.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestViperProvider_OldConfig(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{
		Path: "configs/.rr-old.yaml",
	}

	err := cont.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Foo{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigOverwriteExpandEnv(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = []string{"rpc.listen=tcp://${RPC_VAL:-127.0.0.1:6001}"}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.NoError(t, err)
	_ = container.Stop()
}

func TestConfigOverwriteFail(t *testing.T) {
	container := endure.New(slog.LevelDebug)
	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = []string{"rpc.listen=tcp//not_exist"}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_2(t *testing.T) {
	container := endure.New(slog.LevelDebug)
	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = []string{"rpc.listen="}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_3(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = []string{"="}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteValid(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = []string{"rpc.listen=tcp://127.0.0.1:36643"}

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	err := os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:36643")
	assert.NoError(t, err)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env.yaml"

	err = cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables2(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env2.yaml"

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo4{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables3(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	_ = os.Setenv("RPC_PORT", "6001")

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env3.yaml"

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo5{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariablesFail(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	err := os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:6065")
	assert.NoError(t, err)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env.yaml"

	err = container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.Error(t, err)
}

func TestConfigProvider_GeneralSection(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = cfgPath
	vp.Flags = nil
	vp.Timeout = time.Second * 10

	err := cont.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Foo3{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestEnvArr(t *testing.T) {
	require.NoError(t, syscall.Setenv("REDIS_HOST_1", "localhost:2999"))
	require.NoError(t, syscall.Setenv("REDIS_HOST_2", "localhost:2998"))
	p := &configImpl.Plugin{
		Path:    "configs/.rr-env-arr.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.NoError(t, err)

	str := p.Get("redis.addrs")
	if _, ok := str.([]string); !ok {
		t.Fatal("not a slice")
	}

	require.Len(t, str.([]string), 2)

	if str.([]string)[0] != "localhost:2999" && str.([]string)[0] != "localhost:2998" {
		t.Fatalf("not expanded")
	}

	if str.([]string)[1] != "localhost:2999" && str.([]string)[1] != "localhost:2998" {
		t.Fatalf("not expanded")
	}
}

func TestVersions(t *testing.T) {
	// rr 2.10, config no version
	p := &configImpl.Plugin{
		Path:    "configs/.rr-no-version.yaml",
		Version: "2.10",
	}

	err := p.Init()
	require.Error(t, err)
}

func TestIncludingConfigs(t *testing.T) {
	p := &configImpl.Plugin{
		ExperimentalFeatures: true,
		Path:                 "configs/.rr.include1.yaml",
		Version:              "2023.3.5",
	}

	err := p.Init()
	require.NoError(t, err)

	// Values from .rr.include1.yaml (root file)
	val := p.Get("reload.patterns")
	_ = val
	assert.Equal(t, []any{".php"}, p.Get("reload.patterns"))

	// Values from .rr.include1-sub1.yaml (1st included file)
	assert.Equal(t, "127.0.0.1:15389", p.Get("http.address"))
	assert.Equal(t, "10s", p.Get("reload.interval")) // Checking if overrided

	// Values from .rr.include1-sub2.yaml (2nd included file)
	assert.Equal(t, "console", p.Get("logs.encoding"))
}

func TestErrorWhenIncludedConfigHaveDifferentVersionThenRoot(t *testing.T) {
	p := &configImpl.Plugin{
		ExperimentalFeatures: true,
		Path:                 "configs/.rr.include2.yaml",
		Version:              "2023.3.4",
	}

	err := p.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config_plugin_init: version in included file must be the same as in root")
}

func TestConfigEnvFile(t *testing.T) {
	p := &configImpl.Plugin{
		Path:                 "configs/.rr-env-file.yaml",
		ExperimentalFeatures: true,
		Version:              "2023.3.5",
	}

	err := p.Init()
	require.NoError(t, err)

	// Check if value is get from .env file
	assert.Equal(t, "info", p.Get("logs.level"))

	// Check if included files has also populated values from .env file
	assert.Equal(t, "30s", p.Get("reload.interval"))
}

func TestConfigEnvPriorityWithEnvFile(t *testing.T) {
	err := os.Setenv("LOGS_LEVEL", "debug")
	assert.NoError(t, err)

	p := &configImpl.Plugin{
		Path:                 "configs/.rr-env-file.yaml",
		ExperimentalFeatures: true,
		Version:              "2023.3.5",
	}

	err = p.Init()
	require.NoError(t, err)

	// OS env has higher priority then .env file
	assert.Equal(t, "debug", p.Get("logs.level"))
}
