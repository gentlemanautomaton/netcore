package main

import (
	"github.com/coreos/go-etcd/etcd"
)

func etcdSetup() *etcd.Client {
	etc := etcd.NewClient(nil)
	etc.SetConsistency("WEAK_CONSISTENCY")
	return etc
}
