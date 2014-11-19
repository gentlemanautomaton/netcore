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
	authoritativePool *net.IPNet    // FIXME: actually use this value for something
	guestPool         *net.IPNet    // must be within authoritativePool (at least for now)
	leaseDuration     time.Duration // FIXME: make a separate duration per pool?
	defaultOptions    dhcp4.Options // FIXME: make different options per pool?
	etcdClient        *etcd.Client
}

func dhcpSetup(etc *etcd.Client) chan error {
	etc.CreateDir("dhcp", 0)
	exit := make(chan error, 1)
	serverIP := net.ParseIP("192.168.30.1")
	_, authoritativePool, _ := net.ParseCIDR("192.168.30.0/24")
	guestPool := authoritativePool
	go func() {
		exit <- dhcp4.ListenAndServeIf("vmnet2", &DHCPService{
			ip:                serverIP.To4(),
			leaseDuration:     time.Hour * 12,
			etcdClient:        etc,
			authoritativePool: authoritativePool,
			guestPool:         guestPool,
			defaultOptions: dhcp4.Options{
				dhcp4.OptionSubnetMask:       net.ParseIP("255.255.255.0").To4(), // FIXME: derive this from the authoritativePool?
				dhcp4.OptionRouter:           dhcp4.IPAdd(serverIP, 1).To4(),
				dhcp4.OptionDomainNameServer: serverIP.To4(),
			},
		})
	}()
	return exit
}

// ServeDHCP is called by dhcp4.ListenAndServe when the service is started
func (d *DHCPService) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, reqOptions dhcp4.Options) (response dhcp4.Packet) {
	switch msgType {
	case dhcp4.Discover:
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()
		fmt.Printf("DHCP Discover from %s\n", mac.String())
		ip := d.getIPFromMAC(mac)
		if ip != nil {
			options := d.getOptionsFromMAC(mac)
			fmt.Printf("DHCP Discover from %s (we return %s)\n", mac.String(), ip.String())
			// for x, y := range reqOptions {
			// 	fmt.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	fmt.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip.To4(), ip.To4(), d.leaseDuration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}
		return nil
	case dhcp4.Request:
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()
		fmt.Printf("DHCP Request from %s...\n", mac.String())
		if requestedIP := net.IP(reqOptions[dhcp4.OptionRequestedIPAddress]); len(requestedIP) == 4 { // valid and IPv4
			fmt.Printf("DHCP Request from %s wanting %s\n", mac.String(), requestedIP.String())
			ip := d.getIPFromMAC(mac)
			if ip.Equal(requestedIP) {
				options := d.getOptionsFromMAC(mac)
				fmt.Printf("DHCP Request from %s wanting %s (we agree)\n", mac.String(), requestedIP.String())
				// for x, y := range reqOptions {
				// 	fmt.Printf("\tR[%v] %v %s\n", x, y, y)
				// }
				// for x, y := range options {
				// 	fmt.Printf("\tO[%v] %v %s\n", x, y, y)
				// }
				return dhcp4.ReplyPacket(packet, dhcp4.ACK, d.ip.To4(), requestedIP.To4(), d.leaseDuration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
			}
		}
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
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
	response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/ip", false, false)
	if response != nil && response.Node != nil {
		ip := net.ParseIP(response.Node.Value)
		if ip != nil {
			d.etcdClient.Set("dhcp/"+ip.String(), mac.String(), uint64(d.leaseDuration.Seconds()+0.5))
			return ip
		}
	}

	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)

	// locate an unused IP address (can this be more efficient?  yes!  FIXME)
	var ip net.IP
	for testIP := dhcp4.IPAdd(d.guestPool.IP, 1); d.guestPool.Contains(testIP); testIP = dhcp4.IPAdd(testIP, 1) {
		fmt.Println(testIP.String())
		response, _ := d.etcdClient.Get("dhcp/"+testIP.String(), false, false)
		if response == nil || response.Node == nil { // this means that the IP is not already occupied
			ip = testIP
			break
		}
	}

	if ip != nil { // if nil then we're out of IP addresses!
		d.etcdClient.CreateDir("dhcp/"+mac.String(), 0)
		d.etcdClient.Set("dhcp/"+ip.String(), mac.String(), uint64(d.leaseDuration.Seconds()+0.5))
		d.etcdClient.Set("dhcp/"+mac.String()+"/ip", ip.String(), uint64(d.leaseDuration.Seconds()+0.5))
		return ip
	}

	return nil
}

func (d *DHCPService) getOptionsFromMAC(mac net.HardwareAddr) dhcp4.Options {
	options := dhcp4.Options{}

	for i := range d.defaultOptions {
		options[i] = d.defaultOptions[i]
		fmt.Printf("OPTION:[%d][%+v]\n", i, d.defaultOptions[i])
	}

	{ // Subnet Mask
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/mask", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionSubnetMask)
			} else {
				options[dhcp4.OptionSubnetMask] = []byte(response.Node.Value)
			}
		}
	}

	{ // Gateway/Router
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/gw", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionRouter)
			} else {
				options[dhcp4.OptionRouter] = []byte(response.Node.Value)
			}
		}
	}

	{ // Name Server
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/ns", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionDomainNameServer)
			} else {
				options[dhcp4.OptionDomainNameServer] = []byte(response.Node.Value)
			}
		}
	}

	{ // Host Name
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/name", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionHostName)
			} else {
				options[dhcp4.OptionHostName] = []byte(response.Node.Value)
			}
		}
	}

	{ // Domain Name
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/domain", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionDomainName)
			} else {
				options[dhcp4.OptionDomainName] = []byte(response.Node.Value)
			}
		}
	}

	{ // Broadcast Address
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/broadcast", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionBroadcastAddress)
			} else {
				options[dhcp4.OptionBroadcastAddress] = []byte(response.Node.Value)
			}
		}
	}

	{ // NTP Server
		response, _ := d.etcdClient.Get("dhcp/"+mac.String()+"/ntp", false, false)
		if response != nil && response.Node != nil {
			if response.Node.Value == "" {
				delete(options, dhcp4.OptionNetworkTimeProtocolServers)
			} else {
				options[dhcp4.OptionNetworkTimeProtocolServers] = []byte(response.Node.Value)
			}
		}
	}

	return options
}
