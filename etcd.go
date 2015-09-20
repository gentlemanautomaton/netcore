package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
)

var etcdServers = flag.String("etcd", "", "Comma-separated list of etcd servers.")

// ErrNoEtcdEndpoints indicates that netcore could not find a list of etcd
// endpoints to connect to.
var ErrNoEtcdEndpoints = errors.New("No etcd endpoints provided")

func etcdClient() (client.Client, error) {
	if len(*etcdServers) == 0 {
		if port := os.Getenv("ETCD_PORT"); len(port) > 0 {
			*etcdServers = strings.Replace(port, "tcp://", "http://", 1)
		} else {
			*etcdServers = "etcd" // just some default hostname that Docker or otherwise might use
		}
	}
	if *etcdServers != "" {
		endpoints := strings.Split(*etcdServers, ",")
		return client.New(client.Config{
			Endpoints:               endpoints,
			Transport:               client.DefaultTransport,
			HeaderTimeoutPerRequest: time.Second,
		})
	}
	return nil, ErrNoEtcdEndpoints
}
