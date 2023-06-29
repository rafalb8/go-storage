package etcd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rafalb8/go-storage/internal"
	"github.com/rafalb8/go-storage/internal/iter"
	"github.com/rafalb8/go-storage/internal/net"

	"go.etcd.io/etcd/server/v3/embed"
)

var (
	localIP = net.LocalIP()
)

type etcdEmbed struct {
	cfg   *embed.Config
	embed *embed.Etcd
}

func embedCfg(etcd *Etcd, peers []string, token, dir string, test bool) error {
	cfg := embed.NewConfig()
	// Basic
	cfg.Dir = dir
	nodeIdx, err := thisPeerPosition(peers)
	if err != nil {
		return err
	}
	cfg.Name = "node" + nodeIdx

	// Logger
	cfg.Logger = "zap"
	if test {
		cfg.LogLevel = "error"
	} else {
		cfg.ZapLoggerBuilder = embed.NewZapLoggerBuilder(internal.Logger().Desugar())
	}

	// Peers
	cfg.AdvertisePeerUrls = parseIPs([]string{localIP}, peerURL)
	cfg.ListenPeerUrls = cfg.AdvertisePeerUrls
	cfg.AdvertiseClientUrls = parseIPs([]string{localIP}, clientURL)
	cfg.ListenClientUrls = cfg.AdvertiseClientUrls

	// Cluster cfg
	cfg.InitialClusterToken = token
	cfg.InitialCluster = parseInitialCluster(peers)
	cfg.ClusterState = embed.ClusterStateFlagNew
	if entries, _ := os.ReadDir(cfg.Dir); len(entries) > 0 {
		cfg.ClusterState = embed.ClusterStateFlagExisting
	}

	log.Info("Cluster State:", cfg.ClusterState)
	log.Debug(fmt.Sprintf("%+v", cfg))

	// Set cfg
	etcd.server = &etcdEmbed{cfg: cfg}
	return nil
}

func setupEmbed(etcd *Etcd) error {
	e := etcd.server
	var err error
	e.embed, err = embed.StartEtcd(e.cfg)
	if err != nil {
		return err
	}

	select {
	case <-e.embed.Server.ReadyNotify():
		log.Info("Server is ready!")
		return nil

	case <-time.After(60 * time.Second):
		e.embed.Server.Stop() // trigger a shutdown
		return errors.New("server took too long to start")

	case err := <-e.embed.Err():
		return err
	}
}

func clientURL(ip string) string {
	return "http://" + ip + ":2379"
}

func peerURL(ip string) string {
	return "http://" + ip + ":2380"
}

func thisPeerPosition(peers []string) (string, error) {
	for i, v := range peers {
		if v == localIP {
			return strconv.Itoa(i), nil
		}
	}
	return "", fmt.Errorf("position of %s not found in %s", localIP, peers)
}

func parseIPs(ips []string, format func(string) string) []url.URL {
	log.Debug("Parsing to url.URL:", ips, "format", runtime.FuncForPC(reflect.ValueOf(format).Pointer()).Name())
	return iter.MapSlice(ips, func(ip string) url.URL {
		u, err := url.Parse(format(ip))
		if err != nil {
			return url.URL{}
		}
		return *u
	})
}

func parseInitialCluster(ips []string) string {
	log.Debug("Parsing to initial cluster:", ips)
	i := -1
	return strings.Join(iter.MapSlice(ips, func(ip string) string {
		i++
		return fmt.Sprintf("node%d=%s", i, peerURL(ip))
	}), ",")
}
