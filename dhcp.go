package main

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/krolaw/dhcp4"
)

// DHCPService is the DHCP server instance
type DHCPService struct {
	ip             net.IP
	domain         string
	guestPool      *net.IPNet
	leaseDuration  time.Duration
	defaultOptions dhcp4.Options // FIXME: make different options per pool?
	etcdClient     *etcd.Client
}

type dhcpLease struct {
	mac      net.HardwareAddr
	ip       net.IP
	duration time.Duration
}

const minimumLeaseDuration = 60 * time.Second // FIXME: put this in a config

func dhcpSetup(cfg *Config, etc *etcd.Client) chan error {
	etc.CreateDir("dhcp", 0)
	exit := make(chan error, 1)
	go func() {
		exit <- dhcp4.ListenAndServeIf(cfg.DHCPNIC(), &DHCPService{
			ip:            cfg.DHCPIP(),
			leaseDuration: cfg.DHCPLeaseDuration(),
			etcdClient:    etc,
			guestPool:     cfg.DHCPSubnet(),
			domain:        cfg.Domain(),
			defaultOptions: dhcp4.Options{
				dhcp4.OptionSubnetMask:       net.IP(cfg.Subnet().Mask),
				dhcp4.OptionRouter:           cfg.Gateway(),
				dhcp4.OptionDomainNameServer: cfg.DHCPIP(),
			},
		})
	}()
	return exit
}

// ServeDHCP is called by dhcp4.ListenAndServe when the service is started
func (d *DHCPService) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, reqOptions dhcp4.Options) (response dhcp4.Packet) {
	switch msgType {
	case dhcp4.Discover:
		// RFC 2131 4.3.1
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()

		// Check MAC blacklist
		if !d.isMACPermitted(mac) {
			fmt.Printf("DHCP Discover from %s\n is not permitted", mac.String())
			return nil
		}
		fmt.Printf("DHCP Discover from %s\n", mac.String())

		// Existing Lease
		lease, err := d.getLease(mac)
		if err == nil {
			options := d.getOptionsFromMAC(mac)
			fmt.Printf("DHCP Discover from %s (we offer %s from current lease)\n", mac.String(), lease.ip.String())
			// for x, y := range reqOptions {
			// 	fmt.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	fmt.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip.To4(), lease.ip.To4(), d.getLeaseDurationForRequest(reqOptions, lease.duration), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		// New Lease
		ip := d.getIPFromPool()
		if ip != nil {
			options := d.getOptionsFromMAC(mac)
			fmt.Printf("DHCP Discover from %s (we offer %s from pool)\n", mac.String(), ip.String())
			// for x, y := range reqOptions {
			// 	fmt.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	fmt.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip.To4(), ip.To4(), d.getLeaseDurationForRequest(reqOptions, d.leaseDuration), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		fmt.Printf("DHCP Discover from %s (no offer due to no addresses available in pool)\n", mac.String())
		// FIXME: Send to StatHat and/or increment a counter
		// TODO: Send an email?

		return nil

	case dhcp4.Request:
		// RFC 2131 4.3.2
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()

		// Check MAC blacklist
		if !d.isMACPermitted(mac) {
			fmt.Printf("DHCP Request from %s\n is not permitted", mac.String())
			return nil
		}

		// Check IP presence
		state, requestedIP := d.getRequestState(packet, reqOptions)
		fmt.Printf("DHCP Request (%s) from %s...\n", state, mac.String())
		if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // no IP provided at all... why? FIXME
			fmt.Printf("DHCP Request (%s) from %s (empty IP, so we're just ignoring this request)\n", state, mac.String())
			return nil
		}

		// Check IPv4
		if len(requestedIP) != net.IPv4len {
			fmt.Printf("DHCP Request (%s) from %s wanting %s (IPv6 address requested, so we're just ignoring this request)\n", state, mac.String(), requestedIP.String())
			return nil
		}

		// Check IP subnet
		if !d.guestPool.Contains(requestedIP) {
			fmt.Printf("DHCP Request (%s) from %s wanting %s (we reject due to wrong subnet)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
		}

		// Check Target Server
		targetServerIP := packet.SIAddr()
		if len(targetServerIP) > 0 && !targetServerIP.IsUnspecified() {
			fmt.Printf("DHCP Request (%s) from %s wanting %s is in response to a DHCP offer from %s\n", state, mac.String(), requestedIP.String(), targetServerIP.String())
			if d.ip.Equal(targetServerIP) {
				return nil
			}
		}

		// Process Request
		fmt.Printf("DHCP Request (%s) from %s wanting %s...\n", state, mac.String(), requestedIP.String())
		lease, err := d.getLease(mac)
		if err == nil {
			// Existing Lease
			lease.duration = d.getLeaseDurationForRequest(reqOptions, d.leaseDuration)
			if lease.ip.Equal(requestedIP) {
				err = d.renewLease(lease)
			} else {
				fmt.Printf("DHCP Request (%s) from %s wanting %s (we reject due to lease mismatch, should be %s)\n", state, mac.String(), requestedIP.String(), lease.ip.String())
				return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
			}
		} else {
			// New lease
			lease = dhcpLease{
				mac:      mac,
				ip:       requestedIP,
				duration: d.getLeaseDurationForRequest(reqOptions, d.leaseDuration),
			}
			err = d.createLease(lease)
		}

		if err == nil {
			d.maintainDNSRecords(lease.mac, lease.ip, packet, reqOptions) // TODO: Move this?
			options := d.getOptionsFromMAC(mac)
			fmt.Printf("DHCP Request (%s) from %s wanting %s (we agree)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.ACK, d.ip.To4(), requestedIP.To4(), lease.duration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		fmt.Printf("DHCP Request (%s) from %s wanting %s (we reject due to address collision)\n", state, mac.String(), requestedIP.String())
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)

	case dhcp4.Decline:
		// RFC 2131 4.3.3
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		fmt.Printf("DHCP Decline from %s\n", mac.String())

	case dhcp4.Release:
		// RFC 2131 4.3.4
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		fmt.Printf("DHCP Release from %s\n", mac.String())

	case dhcp4.Inform:
		// RFC 2131 4.3.5
		// https://tools.ietf.org/html/draft-ietf-dhc-dhcpinform-clarify-06
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		// FIXME: we should reply with valuable info, but not assign an IP to this client, per RFC 2131 for DHCPINFORM
		// NOTE: the client's IP is supposed to only be in the ciaddr field, not the requested IP field, per RFC 2131 4.4.3
		mac := packet.CHAddr()
		ip := packet.CIAddr()
		if len(ip) > 0 && !ip.IsUnspecified() {
			fmt.Printf("DHCP Inform from %s for %s \n", mac.String(), ip.String())
			if len(ip) == net.IPv4len && d.guestPool.Contains(ip) {
				options := d.getOptionsFromMAC(mac)
				return informReplyPacket(packet, dhcp4.ACK, d.ip.To4(), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
			}
		}
	}

	return nil
}

func (d *DHCPService) isMACPermitted(mac net.HardwareAddr) bool {
	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)
	return true
}

func (d *DHCPService) getRequestState(packet dhcp4.Packet, reqOptions dhcp4.Options) (string, net.IP) {
	state := "NEW"
	requestedIP := net.IP(reqOptions[dhcp4.OptionRequestedIPAddress])
	if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // empty
		state = "RENEWAL"
		requestedIP = packet.CIAddr()
	}
	return state, requestedIP
}

func (d *DHCPService) getLeaseDurationForRequest(reqOptions dhcp4.Options, defaultDuration time.Duration) time.Duration {
	// If a requested lease duration is accepted by policy we hand it back to them
	// If a requested lease duration is not accepted by policy we constrain it to the policy's minimum and maximum
	// If a lease duration was not requested then we give them the default duration provided to this function
	// The provided default will either be the remaining duration of an existing lease or the configured default duration for the server
	// The provided default will be constrained to the policy's minimum duration
	leaseDuration := defaultDuration

	leaseBytes := reqOptions[dhcp4.OptionIPAddressLeaseTime]
	if len(leaseBytes) == 4 {
		leaseDuration = time.Duration(binary.BigEndian.Uint32(leaseBytes)) * time.Second
		if leaseDuration > d.leaseDuration {
			// The requested lease duration is too long so we give them the maximum allowed by policy
			leaseDuration = d.leaseDuration
		}
	}

	if leaseDuration < minimumLeaseDuration {
		// The lease duration is too short so we give them the minimum allowed by policy
		return minimumLeaseDuration
	}

	return leaseDuration
}

func (d *DHCPService) getIPFromPool() net.IP {
	// locate an unused IP address (can this be more efficient?  yes!  FIXME)
	// TODO: Create a channel and spawn a goproc with something like this function to feed it; then have the server pull addresses from that channel
	for ip := dhcp4.IPAdd(d.guestPool.IP, 1); d.guestPool.Contains(ip); ip = dhcp4.IPAdd(ip, 1) {
		//fmt.Println(ip.String())
		response, _ := d.etcdClient.Get("dhcp/"+ip.String(), false, false)
		if response == nil || response.Node == nil { // this means that the IP is not already occupied
			return ip
		}
	}
	return nil
}

func (d *DHCPService) getLease(mac net.HardwareAddr) (dhcpLease, error) {
	lease := dhcpLease{}
	response, err := d.etcdClient.Get("dhcp/"+mac.String()+"/ip", false, false)
	if err == nil {
		if response != nil && response.Node != nil {
			lease.mac = mac
			lease.ip = net.ParseIP(response.Node.Value)
			lease.duration = time.Duration(response.Node.TTL)
		} else {
			err = errors.New("Not Found")
		}
	}
	return lease, err
}

func (d *DHCPService) renewLease(lease dhcpLease) error {
	duration := uint64(lease.duration.Seconds() + 0.5) // Half second jitter to hide network delay
	_, err := d.etcdClient.CompareAndSwap("dhcp/"+lease.ip.String(), lease.mac.String(), duration, lease.mac.String(), 0)
	if err == nil {
		return d.writeLease(lease)
	}
	return err
}

func (d *DHCPService) createLease(lease dhcpLease) error {
	duration := uint64(lease.duration.Seconds() + 0.5)
	_, err := d.etcdClient.Create("dhcp/"+lease.ip.String(), lease.mac.String(), duration)
	if err == nil {
		return d.writeLease(lease)
	}
	return err
}

func (d *DHCPService) writeLease(lease dhcpLease) error {
	duration := uint64(lease.duration.Seconds() + 0.5) // Half second jitter to hide network delay
	// FIXME: Decide what to do if either of these calls returns an error
	d.etcdClient.CreateDir("dhcp/"+lease.mac.String(), 0)
	d.etcdClient.Set("dhcp/"+lease.mac.String()+"/ip", lease.ip.String(), duration)
	return nil
}

func (d *DHCPService) maintainDNSRecords(mac net.HardwareAddr, ip net.IP, packet dhcp4.Packet, reqOptions dhcp4.Options) {
	options := d.getOptionsFromMAC(mac)
	if domain, ok := options[dhcp4.OptionDomainName]; ok {
		// FIXME:  danger!  we're mixing systems here...  if we keep this up, we will have spaghetti!
		name := ""
		if val, ok := options[dhcp4.OptionHostName]; ok {
			name = string(val)
		} else if val, ok := reqOptions[dhcp4.OptionHostName]; ok {
			name = string(val)
		}
		if name != "" {
			host := strings.ToLower(strings.Join([]string{name, string(domain)}, "."))
			ipHash := fmt.Sprintf("%x", sha1.Sum([]byte(ip.String())))     // hash the IP address so we can have a unique key name (no other reason for this, honestly)
			pathParts := strings.Split(strings.TrimSuffix(host, "."), ".") // breakup the name
			queryPath := strings.Join(reverseSlice(pathParts), "/")        // reverse and join them with a slash delimiter
			fmt.Printf("Wanting to register against %s/%s\n", queryPath, name)
			d.etcdClient.Set("dns/"+queryPath+"/@a/val/"+ipHash, ip.String(), uint64(d.leaseDuration.Seconds()+0.5))
			hostHash := fmt.Sprintf("%x", sha1.Sum([]byte(host))) // hash the hostname so we can have a unique key name (no other reason for this, honestly)
			slashedIP := strings.Replace(ip.To4().String(), ".", "/", -1)
			d.etcdClient.Set("dns/arpa/in-addr/"+slashedIP+"/@ptr/val/"+hostHash, host, uint64(d.leaseDuration.Seconds()+0.5))
		} else {
			fmt.Println(">> No host name")
		}
	} else {
		fmt.Println(">> No domain name")
	}
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
			if response.Node.Value != "" {
				options[dhcp4.OptionDomainName] = []byte(response.Node.Value)
			}
		}
		if len(options[dhcp4.OptionDomainName]) == 0 {
			if d.domain != "" {
				options[dhcp4.OptionDomainName] = []byte(d.domain)
			} else {
				delete(options, dhcp4.OptionDomainName)
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

// ReplyPacket creates a reply packet that a Server would send to a client.
// It uses the req Packet param to copy across common/necessary fields to
// associate the reply with the request.
func informReplyPacket(req dhcp4.Packet, mt dhcp4.MessageType, serverID net.IP, options []dhcp4.Option) dhcp4.Packet {
	p := dhcp4.NewPacket(dhcp4.BootReply)
	p.SetXId(req.XId())
	p.SetHType(req.HType())
	p[2] = req.HLen() // dhcp4 library does not provide a setter
	p.SetFlags(req.Flags())
	p.SetCIAddr(req.CIAddr())
	p.SetCHAddr(req.CHAddr())
	p.AddOption(dhcp4.OptionDHCPMessageType, []byte{byte(mt)})
	p.AddOption(dhcp4.OptionServerIdentifier, []byte(serverID))
	for _, o := range options {
		p.AddOption(o.Code, o.Value)
	}
	p.PadToMinSize()
	return p
}
