package config

type rpc struct {
	pl *Plugin
}

func (r *rpc) Get(name string, output *any) error {
	*output = r.pl.Get(name)

	return nil
}

func (r *rpc) Has(name string, output *bool) error {
	*output = r.pl.Has(name)

	return nil
}
