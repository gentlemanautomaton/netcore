package main

import (
	"fmt"
	"net"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/krolaw/dhcp4"
)

// DHCPService is the DHCP server instance
type DHCPService struct {
	ip                net.IP
	authoritativePool *net.IPNet
	guestPool         *net.IPNet    // must be within authoritativePool (at least for now)
	leaseDuration     time.Duration // FIXME: make a separate duration per pool?
	options           dhcp4.Options // FIXME: make different options per pool?
	etcdClient        *etcd.Client
}

func dhcpSetup(etc *etcd.Client) chan bool {
	etc.CreateDir("dhcp", 0)
	etc.CreateDir("dhcp/mac", 0)
	etc.CreateDir("dhcp/ip", 0)
	exit := make(chan bool, 1)
	serverIP := net.ParseIP("172.16.193.1")
	_, authoritativePool, _ := net.ParseCIDR("172.16.193.0/24")
	guestPool := authoritativePool
	go func() {
		fmt.Println(dhcp4.ListenAndServeIf("vmnet2", &DHCPService{
			ip:                serverIP,
			leaseDuration:     time.Hour * 12,
			etcdClient:        etc,
			authoritativePool: authoritativePool,
			guestPool:         guestPool,
			options: dhcp4.Options{
				dhcp4.OptionSubnetMask:       net.ParseIP("255.255.255.0"),
				dhcp4.OptionRouter:           serverIP,
				dhcp4.OptionDomainNameServer: serverIP,
			},
		}))
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
		fmt.Printf("DHCP Discover from %s\n", mac.String())
		ip := d.getIPFromMAC(mac)
		if ip != nil {
			fmt.Printf("DHCP Discover from %s (we return %s)\n", mac.String(), ip.String())
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip, ip, d.leaseDuration, d.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
		}
		return nil
	case dhcp4.Request:
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()
		fmt.Printf("DHCP Request from %s...\n", mac.String())
		if requestedIP := net.IP(options[dhcp4.OptionRequestedIPAddress]); len(requestedIP) == 4 { // valid and IPv4
			fmt.Printf("DHCP Request from %s wanting %s\n", mac.String(), requestedIP.String())
			ip := d.getIPFromMAC(mac)
			if ip.Equal(requestedIP) {
				return dhcp4.ReplyPacket(packet, dhcp4.ACK, d.ip, requestedIP, d.leaseDuration, d.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
			}
		}
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip, nil, 0, nil)
	case dhcp4.Release:
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		fmt.Printf("DHCP Release from %s\n", mac.String())
	case dhcp4.Decline:
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		fmt.Printf("DHCP Decline from %s\n", mac.String())
	}
	return nil
}

func (d *DHCPService) getIPFromMAC(mac net.HardwareAddr) net.IP {
	response, _ := d.etcdClient.Get("dhcp/mac/"+mac.String(), false, false)
	if response != nil && response.Node != nil {
		ip := net.ParseIP(response.Node.Value)
		if ip != nil {
			d.etcdClient.Set("dhcp/ip/"+ip.String(), mac.String(), uint64(d.leaseDuration.Seconds()+0.5))
			return ip
		}
	}

	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)

	// locate an unused IP address (can this be more efficient?  yes!  FIXME)
	var ip net.IP
	for testIP := dhcp4.IPAdd(d.guestPool.IP, 1); d.guestPool.Contains(testIP); testIP = dhcp4.IPAdd(testIP, 1) {
		fmt.Println(testIP.String())
		response, _ := d.etcdClient.Get("dhcp/ip/"+testIP.String(), false, false)
		if response == nil || response.Node == nil { // this means that the IP is not already occupied
			ip = testIP
			break
		}
	}

	if ip != nil { // if nil then we're out of IP addresses!
		//d.etcdClient.Set("dhcp/ip/"+ip.String(), mac.String(), uint64(d.leaseDuration.Seconds()+0.5))

		// TODO: write this new lease to the database under both "dhcp/mac/$MAC" and a MAC pointer to "dhcp/ip/$IP"
		// TODO: determine dhcp4.options based on server default config + MAC config
		// TODO: return valid IP as expected

		return ip
	}

	return nil
}
