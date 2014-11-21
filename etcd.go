package main

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func etcdSetup(serverList string) *etcd.Client {
	var servers []string
	if serverList != "" {
		servers = strings.Split(serverList, ",")
	}
	etc := etcd.NewClient(servers)
	etc.SetConsistency("WEAK_CONSISTENCY")
	return etc
}

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
