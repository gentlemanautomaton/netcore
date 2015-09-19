package main

import (
	"flag"
	"os"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

var etcdServers = flag.String("etcd", "", "Comma-separated list of etcd servers.")

func etcdClient() etcd.Client {
	if len(*etcdServers) == 0 {
		if len(os.Getenv("ETCD_PORT")) > 0 {
			*etcdServers = strings.Replace(os.Getenv("ETCD_PORT"), "tcp://", "http://", 1)
		} else {
			*etcdServers = "etcd" // just some default hostname that Docker or otherwise might use
		}
	}
	var servers []string
	if etcdServers != "" {
		servers = strings.Split(etcdServers, ",")
	}
	return etcd.NewClient(servers)
}
