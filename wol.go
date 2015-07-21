package main

import (
	"errors"
	"net"
	"strings"

	"github.com/coreos/go-etcd/etcd"
	"github.com/sabhiram/go-wol"
)

func wakeByMAC(e *etcd.Client, mac string) error {
	return wol.SendMagicPacket(mac, "255.255.255.255:9", "")
}

func wakeByIP(e *etcd.Client, ip net.IP) error {
	response, err := e.Get("/dhcp/"+ip.String(), false, false)
	if err != nil {
		return err
	}
	if response == nil || response.Node == nil {
		err = errors.New("Not Found")
	}
	mac := response.Node.Value
	return wakeByMAC(e, mac)
}

func wakeByHostname(e *etcd.Client, hostname string) error {
	pathParts := strings.Split(strings.TrimSuffix(hostname, "."), ".") // breakup the queryed name
	queryPath := strings.Join(reverseSlice(pathParts), "/")            // reverse and join them with a slash delimiter
	keyRoot := strings.ToLower("/dns/" + queryPath)
	response, err := e.Get(keyRoot+"/@a/val", true, true)
	if err == nil {
		if response != nil && response.Node != nil && response.Node.Nodes != nil {
			for _, node := range response.Node.Nodes {
				ip := node.Value
				err = wakeByIP(e, net.ParseIP(ip))
			}
		} else {
			err = errors.New("Not Found")
		}
	}
	return err
}
