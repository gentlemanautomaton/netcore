package main

import (
	"encoding/binary"
	"log"
	"net"
	"strings"
	"time"

	"github.com/krolaw/dhcp4"
)

type DHCPDB interface {
	InitDHCP()
	GetIP(net.IP) (IPEntry, error)
	HasIP(net.IP) bool
	GetMAC(mac net.HardwareAddr, cascade bool) (entry *MACEntry, found bool, err error)
	RenewLease(lease *MACEntry) error
	CreateLease(lease *MACEntry) error
	WriteLease(lease *MACEntry) error
}

// DHCPService is the DHCP server instance
type DHCPService struct {
	ip             net.IP
	domain         string
	subnet         *net.IPNet
	guestPool      *net.IPNet
	leaseDuration  time.Duration
	defaultOptions dhcp4.Options // FIXME: make different options per pool?
	db             DB
}

type IPEntry struct {
	MAC net.HardwareAddr
}

type MACEntry struct {
	MAC      net.HardwareAddr
	IP       net.IP
	Duration time.Duration
	Attr     map[string]string
}

const minimumLeaseDuration = 60 * time.Second // FIXME: put this in a config

func dhcpSetup(cfg *Config) chan error {
	cfg.db.InitDHCP()
	exit := make(chan error, 1)
	go func() {
		d := &DHCPService{
			ip:            cfg.DHCPIP(),
			leaseDuration: cfg.DHCPLeaseDuration(),
			db:            cfg.db,
			subnet:        cfg.Subnet(),
			guestPool:     cfg.DHCPSubnet(),
			domain:        cfg.Domain(),
			defaultOptions: dhcp4.Options{
				dhcp4.OptionSubnetMask:       net.IP(cfg.Subnet().Mask),
				dhcp4.OptionRouter:           cfg.Gateway(),
				dhcp4.OptionDomainNameServer: cfg.DHCPIP(),
			},
		}
		dhcpTFTP := cfg.DHCPTFTP()
		if dhcpTFTP != "" {
			d.defaultOptions[dhcp4.OptionTFTPServerName] = []byte(dhcpTFTP)
		}
		exit <- dhcp4.ListenAndServeIf(cfg.DHCPNIC(), d)
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
			log.Printf("DHCP Discover from %s\n is not permitted", mac.String())
			return nil
		}
		log.Printf("DHCP Discover from %s\n", mac.String())

		// Look up the MAC entry with cascaded attributes
		lease, found, err := d.db.GetMAC(mac, true)
		if err != nil {
			return nil
		}

		// Existing Lease
		if found {
			options := d.getOptionsFromMAC(lease)
			log.Printf("DHCP Discover from %s (we offer %s from current lease)\n", lease.MAC.String(), lease.IP.String())
			// for x, y := range reqOptions {
			// 	log.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	log.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip.To4(), lease.IP.To4(), d.getLeaseDurationForRequest(reqOptions, lease.Duration), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		// New Lease
		ip := d.getIPFromPool()
		if ip != nil {
			options := d.getOptionsFromMAC(lease)
			log.Printf("DHCP Discover from %s (we offer %s from pool)\n", mac.String(), ip.String())
			// for x, y := range reqOptions {
			// 	log.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	log.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, d.ip.To4(), ip.To4(), d.getLeaseDurationForRequest(reqOptions, d.leaseDuration), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		log.Printf("DHCP Discover from %s (no offer due to no addresses available in pool)\n", mac.String())
		// FIXME: Send to StatHat and/or increment a counter
		// TODO: Send an email?

		return nil

	case dhcp4.Request:
		// RFC 2131 4.3.2
		// FIXME: send to StatHat and/or increment a counter
		mac := packet.CHAddr()

		// Check MAC blacklist
		if !d.isMACPermitted(mac) {
			log.Printf("DHCP Request from %s\n is not permitted", mac.String())
			return nil
		}

		// Check IP presence
		state, requestedIP := d.getRequestState(packet, reqOptions)
		log.Printf("DHCP Request (%s) from %s...\n", state, mac.String())
		if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // no IP provided at all... why? FIXME
			log.Printf("DHCP Request (%s) from %s (empty IP, so we're just ignoring this request)\n", state, mac.String())
			return nil
		}

		// Check IPv4
		if len(requestedIP) != net.IPv4len {
			log.Printf("DHCP Request (%s) from %s wanting %s (IPv6 address requested, so we're just ignoring this request)\n", state, mac.String(), requestedIP.String())
			return nil
		}

		// Check IP subnet
		if !d.subnet.Contains(requestedIP) {
			log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to wrong subnet)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
		}

		// Check Target Server
		targetServerIP := packet.SIAddr()
		if len(targetServerIP) > 0 && !targetServerIP.IsUnspecified() {
			log.Printf("DHCP Request (%s) from %s wanting %s is in response to a DHCP offer from %s\n", state, mac.String(), requestedIP.String(), targetServerIP.String())
			if d.ip.Equal(targetServerIP) {
				return nil
			}
		}

		// Process Request
		log.Printf("DHCP Request (%s) from %s wanting %s...\n", state, mac.String(), requestedIP.String())
		lease, found, err := d.db.GetMAC(mac, true)
		if err != nil {
			return nil
		}

		if found {
			// Existing Lease
			lease.Duration = d.getLeaseDurationForRequest(reqOptions, d.leaseDuration)
			if lease.IP.Equal(requestedIP) {
				err = d.db.RenewLease(lease)
			} else {
				log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to lease mismatch, should be %s)\n", state, lease.MAC.String(), requestedIP.String(), lease.IP.String())
				return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
			}
		} else {
			// Check IP subnet is within the guestPool (we don't want users requesting non-pool addresses unless we assigned it to their MAC, administratively)
			if !d.guestPool.Contains(requestedIP) {
				log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to not being within the guestPool)\n", state, mac.String(), requestedIP.String())
				return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)
			}

			// New lease
			lease = &MACEntry{
				MAC:      mac,
				IP:       requestedIP,
				Duration: d.getLeaseDurationForRequest(reqOptions, d.leaseDuration),
			}
			err = d.db.CreateLease(lease)
		}

		if err == nil {
			d.maintainDNSRecords(lease, packet, reqOptions) // TODO: Move this?
			options := d.getOptionsFromMAC(lease)
			log.Printf("DHCP Request (%s) from %s wanting %s (we agree)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.ACK, d.ip.To4(), requestedIP.To4(), lease.Duration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to address collision)\n", state, mac.String(), requestedIP.String())
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, d.ip.To4(), nil, 0, nil)

	case dhcp4.Decline:
		// RFC 2131 4.3.3
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		log.Printf("DHCP Decline from %s\n", mac.String())

	case dhcp4.Release:
		// RFC 2131 4.3.4
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		mac := packet.CHAddr()
		log.Printf("DHCP Release from %s\n", mac.String())

	case dhcp4.Inform:
		// RFC 2131 4.3.5
		// https://tools.ietf.org/html/draft-ietf-dhc-dhcpinform-clarify-06
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		// FIXME: we should reply with valuable info, but not assign an IP to this client, per RFC 2131 for DHCPINFORM
		// NOTE: the client's IP is supposed to only be in the ciaddr field, not the requested IP field, per RFC 2131 4.4.3
		mac := packet.CHAddr()
		ip := packet.CIAddr()
		if len(ip) > 0 && !ip.IsUnspecified() {
			log.Printf("DHCP Inform from %s for %s \n", mac.String(), ip.String())
			if len(ip) == net.IPv4len && d.guestPool.Contains(ip) {
				entry, found, _ := d.db.GetMAC(mac, true)
				if found {
					options := d.getOptionsFromMAC(entry)
					return informReplyPacket(packet, dhcp4.ACK, d.ip.To4(), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
				}
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
		//log.Println(ip.String())
		if !d.db.HasIP(ip) { // this means that the IP is not already occupied
			return ip
		}
	}
	return nil
}

func (d *DHCPService) maintainDNSRecords(entry *MACEntry, packet dhcp4.Packet, reqOptions dhcp4.Options) {
	options := d.getOptionsFromMAC(entry)
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
			// TODO: Pick a TTL for the record and use it
			d.db.RegisterA(host, entry.IP, false, 0, uint64(d.leaseDuration.Seconds()+0.5))
		} else {
			log.Println(">> No host name")
		}
	} else {
		log.Println(">> No domain name")
	}
}

func (d *DHCPService) getOptionsFromMAC(entry *MACEntry) dhcp4.Options {
	options := dhcp4.Options{}

	for i := range d.defaultOptions {
		options[i] = d.defaultOptions[i]
		log.Printf("OPTION:[%d][%+v]\n", i, d.defaultOptions[i])
	}

	{ // Subnet Mask
		if value, ok := entry.Attr["mask"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionSubnetMask)
			} else {
				options[dhcp4.OptionSubnetMask] = []byte(value)
			}
		}
	}

	{ // Gateway/Router
		if value, ok := entry.Attr["gw"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionRouter)
			} else {
				options[dhcp4.OptionRouter] = []byte(value)
			}
		}
	}

	{ // Name Server
		if value, ok := entry.Attr["ns"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionDomainNameServer)
			} else {
				options[dhcp4.OptionDomainNameServer] = []byte(value)
			}
		}
	}

	{ // Host Name
		if value, ok := entry.Attr["name"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionHostName)
			} else {
				options[dhcp4.OptionHostName] = []byte(value)
			}
		}
	}

	{ // Domain Name
		if value, ok := entry.Attr["domain"]; ok {
			if value != "" {
				options[dhcp4.OptionDomainName] = []byte(value)
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
		if value, ok := entry.Attr["broadcast"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionBroadcastAddress)
			} else {
				options[dhcp4.OptionBroadcastAddress] = []byte(value)
			}
		}
	}

	{ // NTP Server
		if value, ok := entry.Attr["ntp"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionNetworkTimeProtocolServers)
			} else {
				options[dhcp4.OptionNetworkTimeProtocolServers] = []byte(value)
			}
		}
	}

	{ // TFTP Server
		if value, ok := entry.Attr["tftp"]; ok {
			if value == "" {
				delete(options, dhcp4.OptionTFTPServerName)
			} else {
				options[dhcp4.OptionTFTPServerName] = []byte(value)
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
