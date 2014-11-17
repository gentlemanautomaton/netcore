package main

import (
	"github.com/coreos/go-etcd/etcd"
)

func etcdSetup() *etcd.Client {
	etc := etcd.NewClient(nil)
	etc.SetConsistency("WEAK_CONSISTENCY")
	etc.CreateDir("dhcp", 0)
	return etc
}
