package main

import (
	"net"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/krolaw/dhcp4"
)

// DHCPService is the DHCP server instance
type DHCPService struct {
	ip                net.IP
	authoritativePool net.IPNet
	guestPool         net.IPNet     // must be within authoritativePool (at least for now)
	leaseDuration     time.Duration // FIXME: make a separate duration per pool?
	options           dhcp4.Options // FIXME: make different options per pool?
	etcdClient        *etcd.Client
}

func dhcpSetup(etc *etcd.Client) chan bool {
	etc.CreateDir("dhcp", 0)
	etc.CreateDir("dhcp/mac", 0)
	etc.CreateDir("dhcp/ip", 0)
	exit := make(chan bool, 1)
	go func() {
		dhcp4.ListenAndServe(&DHCPService{
			ip:            net.ParseIP("10.100.0.121"),
			leaseDuration: time.Hour * 12,
			etcdClient:    etc,
			options: dhcp4.Options{
				dhcp4.OptionSubnetMask:       net.ParseIP("255.255.252.0"),
				dhcp4.OptionRouter:           net.ParseIP("10.100.0.1"),
				dhcp4.OptionDomainNameServer: net.ParseIP("10.100.1.1"),
			},
		})
		exit <- true
	}()
	return exit
}

// ServeDHCP is called by dhcp4.ListenAndServe when the service is started
func (d *DHCPService) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (response dhcp4.Packet) {
	switch msgType {
	case dhcp4.Discover:
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()
		ip := d.getIPFromMAC(mac)
		if ip != nil {
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip, ip, d.leaseDuration, d.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
		}
		return nil
	case dhcp4.Request:
		// FIXME: send to StatHat and/or increment a counter
		if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(d.ip) {
			return nil // not directed at us, so let's ignore it.
		}
		if requestedIP := net.IP(options[dhcp4.OptionRequestedIPAddress]); len(requestedIP) == 4 { // valid and IPv4
			mac := packet.CHAddr()
			ip := d.getIPFromMAC(mac)
			if ip.Equal(requestedIP) {
				return dhcp4.ReplyPacket(packet, dhcp4.ACK, d.ip, requestedIP, d.leaseDuration, d.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
			}
		}
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip, nil, 0, nil)
	case dhcp4.Release:
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
	case dhcp4.Decline:
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
	}
	return nil
}

func (d *DHCPService) getIPFromMAC(mac net.HardwareAddr) net.IP {
	response, _ := d.etcdClient.Get("dhcp/mac/"+mac.String(), false, false)
	ip := net.ParseIP(response.Node.Value)
	if ip != nil {
		d.etcdClient.Set("dhcp/ip/"+ip.String(), mac.String(), uint64(d.leaseDuration.Seconds()+0.5))
		return ip
	}

	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)
	// TODO: locate an unused IP address
	// TODO: write this new lease to the database under both "dhcp/mac/$MAC" and a MAC pointer to "dhcp/ip/$IP"
	// TODO: determine dhcp4.options based on server default config + MAC config
	// TODO: return valid IP as expected

	return net.ParseIP("10.100.3.254") // FIXME: instead of this obviously wrong action, generate (and store!) an IP for this MAC address
}
