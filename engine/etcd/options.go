package etcd

import (
	"context"

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
		log.Info("Found peers", peers)

		// Add self to ips
		peers = append(peers, localIP)

		e.endpoints = parseClusterClients(peers)

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

func parseClusterClients(ips []string) []string {
	log.Debug("Parsing cluster clients:", ips)
	return iter.MapSlice(ips, func(ip string) string { return clientURL(ip) })
}
