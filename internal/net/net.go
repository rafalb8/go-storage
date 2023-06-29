package net

import (
	"context"
	"net"
	"sort"
	"time"
)

const DNSTimeout = 100 * time.Millisecond

func DNSResolve(url string) []string {
	if url == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DNSTimeout)
	defer cancel()

	var r net.Resolver

	addrs, err := r.LookupIP(ctx, "ip4", url)
	if err != nil {
		return nil
	}

	ips := make([]string, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.String()
	}
	sort.Strings(ips)
	return ips
}

// LocalIP returns the non loopback local IP of the host
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
