package main

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type EtcdDB struct {
	client *etcd.Client
}

func NewEtcdDB(serverList string) DB {
	var servers []string
	if serverList != "" {
		servers = strings.Split(serverList, ",")
	}
	client := etcd.NewClient(servers)
	client.SetConsistency("WEAK_CONSISTENCY")
	db := EtcdDB{client}
	return db
}

func etcdKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Key not found")
}
