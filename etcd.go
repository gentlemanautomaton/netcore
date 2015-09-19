package main

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

var etcdServers = flag.String("etcd", "", "Comma-separated list of etcd servers.")

var ErrNoEtcdServers = errors.New("No etcd server list provided")

func etcdClient() (*etcd.Client, error) {
	if len(*etcdServers) == 0 {
		if port := os.Getenv("ETCD_PORT"); len(port) > 0 {
			*etcdServers = strings.Replace(port, "tcp://", "http://", 1)
		} else {
			*etcdServers = "etcd" // just some default hostname that Docker or otherwise might use
		}
	}
	var servers []string
	if *etcdServers != "" {
		servers = strings.Split(*etcdServers, ",")
		return etcd.NewClient(servers), nil
	}
	return nil, ErrNoEtcdServers
}
