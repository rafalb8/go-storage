package etcd

import (
	"context"

	"github.com/rafalb8/go-storage"
	"github.com/rafalb8/go-storage/encoding"
	"github.com/rafalb8/go-storage/internal/iter"
	"github.com/rafalb8/go-storage/internal/net"
)

type EtcdOpts func(*Etcd) error

// Connect to etcd endpoints
func Endpoints(edp ...string) EtcdOpts {
	return func(e *Etcd) error {
		e.endpoints = edp
		return nil
	}
}

// Start single-node embedded etcd
func Embed(loadBalancer, token, dir string, test bool) EtcdOpts {
	return func(e *Etcd) error {
		peers := net.DNSResolve(loadBalancer)

		// Add self to ips
		peers = append(peers, localIP)

		e.endpoints = iter.MapSlice(peers, func(ip string) string { return clientURL(ip) })

		err := embedCfg(e, peers, token, dir, test)
		if err != nil {
			return err
		}
		return setupEmbed(e)
	}
}

func Coder(coder encoding.Coder) EtcdOpts {
	return func(e *Etcd) error {
		e.encoding = coder
		return nil
	}
}

func Context(ctx context.Context) EtcdOpts {
	return func(e *Etcd) error {
		e.ctx, e.cancel = context.WithCancel(ctx)
		return nil
	}
}

func Logger(lg storage.Logger) EtcdOpts {
	return func(e *Etcd) error {
		e.lg = lg
		return nil
	}
}
