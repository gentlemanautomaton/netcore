package netdhcpetcd

import "context"

// Global returns global configuration data stored in etcd.
func (p *Provider) Global(ctx context.Context) (Global, error) {
	//p.c.Get(ctx, )
}

// Instance returns instance configuration data stored in etcd.
func (p *Provider) Instance(ctx context.Context, id string) (Instance, error) {
}
