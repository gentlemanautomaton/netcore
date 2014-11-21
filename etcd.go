package main

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

func etcdSetup() *etcd.Client {
	etc := etcd.NewClient(nil)
	etc.SetConsistency("WEAK_CONSISTENCY")
	return etc
}

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
