package netdhcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/krolaw/dhcp4"
)

// Service provides netcore DHCP services.
type Service struct {
	m       sync.RWMutex // guards access to v
	v       *service     // current view of the service, protected by m
	started chan bool
	done    chan Completion
}

// NewService creates a new netcore DHCP service with the given data provider
// service instance ID.
func NewService(provider Provider, instance string) *Service {
	s := &Service{
		started: make(chan bool, 1),
		done:    make(chan Completion, 1),
	}

	go s.init(provider, instance)

	return s
}

func (s *Service) init(p Provider, id string) {
	if err := s.load(p, id); err != nil {
		s.signalStarted(false)
		s.signalDone(false, err)
		return
	}
	s.signalStarted(true)
	err := dhcp4.ListenAndServeIf(s.view().cfg.NIC(), s)
	s.signalDone(true, err)
}

func (s *Service) signalStarted(success bool) {
	s.started <- success
	close(s.started)
}

func (s *Service) signalDone(initialized bool, err error) {
	s.done <- Completion{false, err}
	close(s.done)
}

func (s *Service) load(p Provider, id string) error {
	if id == "" {
		return errors.New("No instance identifier")
	}

	if p.GlobalProvider == nil {
		return errors.New("No global provider")
	}
	if p.NetworkProvider == nil {
		return errors.New("No network provider")
	}
	if p.InstanceProvider == nil {
		return errors.New("No instance provider")
	}
	// FIXME: Check for all required provider interfaces

	// TODO: Create watcher for each component and reload when configuration changes.

	global := NewContext(p)
	instance := global.Instance(id)

	ctx := context.Background() // FIXME: Use a real cancellable context

	globalData, err := global.Read(ctx)
	if err != nil {
		return err
	}

	instanceData, err := instance.Read(ctx)
	if err != nil {
		return err
	}

	var netID string
	switch {
	case instanceData.Network != "":
		netID = instanceData.Network
	case globalData.Network != "":
		netID = globalData.Network
	default:
		return ErrNoConfigNetwork
	}

	network := global.Network(netID)
	networkData, err := network.Read(ctx)
	if err != nil {
		return err
	}

	cfg := Merge(globalData, instanceData, networkData) // TODO: Review order
	if err := Validate(cfg); err != nil {
		return err
	}

	srv := &service{
		id:       id,
		global:   global,
		instance: instance,
		network:  network,
		cfg:      cfg,
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.v = srv

	//s.inst = inst
	//s.net = netwrk

	return nil
}

// view returns a consistent view of the service.
func (s *Service) view() *service {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.v
}

type params struct {
	p  Provider
	id string
}

func watch(ctx context.Context, p Provider, id string, gs chan GlobalResult, is chan InstanceResult, ns chan NetworkResult) {
	opt := WatcherOptions{}
	global, instance, network := load(p, id)
	gw := global.Watcher(opt)
	iw := instance.Watcher(opt)
	nw := network.Watcher(opt)
	go func() {
		g, err := gw.Next(ctx)
		gs <- GlobalResult{g, err}
	}()
	go func() {
		i, err := iw.Next(ctx)
		is <- InstanceResult{i, err}
	}()
	go func() {
		n, err := nw.Next(ctx)
		ns <- NetworkResult{n, err}
	}()
}

func (s *Service) sync(chan params) {
	// TODO: Figure this out
	var (
		gs       chan GlobalResult
		is       chan InstanceResult
		ns       chan NetworkResult
		global   GlobalContext
		instance InstanceContext
		network  NetworkContext
	)
	for {
		select {
		case input <- params:
			global = NewContext(input.p)
			instance = global.Instance(input.id)
			wctx := context.Background()
			go func() {
				g, err := gw.Next(wctx)
				gs <- GlobalResult{g, err}
			}()
			go func() {
				i, err := iw.Next(wctx)
				is <- InstanceResult{i, err}
			}()
			go func() {
				n, err := nw.Next(wctx)
				ns <- NetworkResult{n, err}
			}()
		case <-gs:
		case <-is:
		case <-ns:
		}
	}
	for {
		v := s.view()
		opt := WatcherOptions{}
		gw := v.global.Watcher(opt)
		iw := v.instance.Watcher(opt)
		nw := v.network.Watcher(opt)
		var gs, is, ns chan struct{} // Signals
		go func() {
			gw.Next(ctx)
			gs <- struct{}{}
		}()
		go func() {
			iw.Next(ctx)
			is <- struct{}{}
		}()
		go func() {
			nw.Next(ctx)
			ns <- struct{}{}
		}()
		select {
		case <-gs:
		case <-is:
		case <-ns:
		}
		// TODO: Signal the ctx?
		// TODO: Reload
	}
}

// Started returns a channel that will be signaled when the service has started
// or failed to start. If the returned value is true the service started
// succesfully.
func (s *Service) Started() chan bool {
	return s.started
}

// Done returns a channel that will be signaled when the service exits.
func (s *Service) Done() chan Completion {
	return s.done
}

// ServeDHCP is called by dhcp4.ListenAndServe when the service is started
func (s *Service) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, reqOptions dhcp4.Options) (response dhcp4.Packet) {
}

type cache struct {
	global   Global
	instance Instance
	network  Network
}

// service operates within a consistent cached view of global, instance and
// network configuration.
type service struct {
	id       string
	global   GlobalContext
	instance InstanceContext
	network  NetworkContext
	cfg      Config
	//cache    cache
}

func (s *service) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, reqOptions dhcp4.Options) (response dhcp4.Packet) {
	switch msgType {
	case dhcp4.Discover:
		// RFC 2131 4.3.1
		// FIXME: send to StatHat and/or increment a counter
		addr := packet.CHAddr()

		// Check MAC blacklist
		if !s.isMACPermitted(addr) {
			log.Printf("DHCP Discover from %s\n is not permitted", addr.String())
			return nil
		}
		log.Printf("DHCP Discover from %s\n", addr.String())

		// FIXME: Look up the MAC entry with cascaded attributes?
		data, found, err := s.net.MAC.Lookup(context.Background(), addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		ip, lease, err := s.selectIP(context.Background(), data.IP, addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		for _, prefix := range macPrefixes(addr) {

		}

		// Existing Lease
		if found {
			options := s.getOptionsFromMAC(data)
			log.Printf("DHCP Discover from %s (we offer %s from current lease)\n", addr.String(), lease.IP.String())
			// for x, y := range reqOptions {
			// 	log.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	log.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, s.cfg.IP().To4(), lease.IP.To4(), s.getLeaseDurationForRequest(reqOptions, lease.Duration), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		// New Lease
		ip := s.getIPFromPool()
		if ip != nil {
			options := s.getOptionsFromMAC(lease)
			log.Printf("DHCP Discover from %s (we offer %s from pool)\n", mac.String(), ip.String())
			// for x, y := range reqOptions {
			// 	log.Printf("\tR[%v] %v %s\n", x, y, y)
			// }
			// for x, y := range options {
			// 	log.Printf("\tO[%v] %v %s\n", x, y, y)
			// }
			return dhcp4.ReplyPacket(packet, dhcp4.Offer, s.cfg.IP().To4(), ip.To4(), s.getLeaseDurationForRequest(reqOptions, s.cfg.LeaseDuration()), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
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
		if !s.isMACPermitted(mac) {
			log.Printf("DHCP Request from %s\n is not permitted", mac.String())
			return nil
		}

		// Check IP presence
		state, requestedIP := s.getRequestState(packet, reqOptions)
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
		if !s.cfg.Subnet().Contains(requestedIP) {
			log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to wrong subnet)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.cfg.IP().To4(), nil, 0, nil)
		}

		// Check Target Server
		targetServerIP := packet.SIAddr()
		if len(targetServerIP) > 0 && !targetServerIP.IsUnspecified() {
			log.Printf("DHCP Request (%s) from %s wanting %s is in response to a DHCP offer from %s\n", state, mac.String(), requestedIP.String(), targetServerIP.String())
			if s.cfg.IP().Equal(targetServerIP) {
				return nil
			}
		}

		// Process Request
		log.Printf("DHCP Request (%s) from %s wanting %s...\n", state, mac.String(), requestedIP.String())
		lease, found, err := s.p.MAC(mac, true)
		if err != nil {
			return nil
		}

		if found {
			// Existing Lease
			lease.Duration = s.getLeaseDurationForRequest(reqOptions, s.cfg.LeaseDuration())
			if lease.IP.Equal(requestedIP) {
				err = s.p.RenewLease(lease)
			} else {
				log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to lease mismatch, should be %s)\n", state, lease.MAC.String(), requestedIP.String(), lease.IP.String())
				return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.cfg.IP().To4(), nil, 0, nil)
			}
		} else {
			// Check IP subnet is within the guestPool (we don't want users requesting non-pool addresses unless we assigned it to their MAC, administratively)
			if !s.cfg.GuestPool().Contains(requestedIP) {
				log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to not being within the guestPool)\n", state, mac.String(), requestedIP.String())
				return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.cfg.IP().To4(), nil, 0, nil)
			}

			// New lease
			lease = &MACEntry{
				MAC:      mac,
				IP:       requestedIP,
				Duration: s.getLeaseDurationForRequest(reqOptions, s.cfg.LeaseDuration()),
			}
			err = s.p.CreateLease(lease)
		}

		if err == nil {
			s.maintainDNSRecords(lease, packet, reqOptions) // TODO: Move this?
			options := s.getOptionsFromMAC(lease)
			log.Printf("DHCP Request (%s) from %s wanting %s (we agree)\n", state, mac.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.ACK, s.cfg.IP().To4(), requestedIP.To4(), lease.Duration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
		}

		log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to address collision)\n", state, mac.String(), requestedIP.String())
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.cfg.IP().To4(), nil, 0, nil)

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
			if len(ip) == net.IPv4len && s.cfg.GuestPool().Contains(ip) {
				entry, found, _ := s.p.MAC(mac, true)
				if found {
					options := s.getOptionsFromMAC(entry)
					return informReplyPacket(packet, dhcp4.ACK, s.cfg.IP().To4(), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
				}
			}
		}
	}

	return nil
}

func (s *Service) isMACPermitted(mac net.HardwareAddr) bool {
	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)
	return true
}

func (s *Service) pepareLease(target net.HardwareAddr) {

}

func (s *Service) selectIP(ctx context.Context, ipset []*IP, target net.HardwareAddr) (*IP, *Lease, error) {
	// Enumerate all reservations and dynamic allocations and select the most
	// appropriate lease based on the following algorithm:
	// 1. Reservation with active lease matching this MAC, sorted by priority and then by recency
	// 2. Reservation without any active lease, sorted by priority and then by recency
	// 3. Dynamic Allocation with active lease matching this MAC, sorted by priority and then by recency
	// 4. Dynamic Allocation without any active lease, sorted by priority and then by recency
	// 5. New Dynamic Allocation from pool

	// TODO: Consider whether we should always give the lease of highest priority
	//       regardless of whatever the current lease is.

	type result struct {
		IP    *IP
		Lease *Lease
		Found bool
		Err   error
	}

	if len(ipset) == 0 {
		return nil, nil, nil
	}

	sort.Sort(IPSet(ipset))

	// Issue a lease lookup for each potential IP address in parallel
	queries := make([]chan result, 0, len(ipset))
	for _, ip := range ipset {
		// FIXME: Make sure the ip address is within this network
		ch := make(chan leaseLookup)
		queries = append(queries, ch)
		go func(ip *IP) {
			lease, found, err := s.net.Lease.Lookup(ctx, ip.Address)
			ch <- result{lease, found, err}
		}(ip)
	}

	// Process each ip and it's resulting lease lookup in preferential order
	for _, result := range queries {
		if result.Err != nil {
			return nil, result.Err
		}
		if result.Found {
			if bytes.Equal(result.Lease.MAC, target) {
				return ipset[i], result.Lease, nil
			}
		}
		return result.IP, nil, nil
	}
	return nil, nil, nil
}

func (s *Service) getRequestState(packet dhcp4.Packet, reqOptions dhcp4.Options) (string, net.IP) {
	state := "NEW"
	requestedIP := net.IP(reqOptions[dhcp4.OptionRequestedIPAddress])
	if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // empty
		state = "RENEWAL"
		requestedIP = packet.CIAddr()
	}
	return state, requestedIP
}

func (s *Service) getLeaseDurationForRequest(reqOptions dhcp4.Options, defaultDuration time.Duration) time.Duration {
	// If a requested lease duration is accepted by policy we hand it back to them
	// If a requested lease duration is not accepted by policy we constrain it to the policy's minimum and maximum
	// If a lease duration was not requested then we give them the default duration provided to this function
	// The provided default will either be the remaining duration of an existing lease or the configured default duration for the server
	// The provided default will be constrained to the policy's minimum duration
	leaseDuration := defaultDuration

	leaseBytes := reqOptions[dhcp4.OptionIPAddressLeaseTime]
	if len(leaseBytes) == 4 {
		leaseDuration = time.Duration(binary.BigEndian.Uint32(leaseBytes)) * time.Second
		if leaseDuration > s.cfg.LeaseDuration() {
			// The requested lease duration is too long so we give them the maximum allowed by policy
			leaseDuration = s.cfg.LeaseDuration()
		}
	}

	if leaseDuration < minimumLeaseDuration {
		// The lease duration is too short so we give them the minimum allowed by policy
		return minimumLeaseDuration
	}

	return leaseDuration
}

func (s *Service) getIPFromPool() net.IP {
	// locate an unused IP address (can this be more efficient?  yes!  FIXME)
	// TODO: Create a channel and spawn a goproc with something like this function to feed it; then have the server pull addresses from that channel
	gp := s.cfg.GuestPool()
	for ip := dhcp4.IPAdd(gp.IP, 1); gp.Contains(ip); ip = dhcp4.IPAdd(ip, 1) {
		//log.Println(ip.String())
		if !s.p.HasIP(ip) { // this means that the IP is not already occupied
			return ip
		}
	}
	return nil
}

func (s *Service) maintainDNSRecords(entry *MACEntry, packet dhcp4.Packet, reqOptions dhcp4.Options) {
	options := s.getOptionsFromMAC(entry)
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
			s.p.RegisterA(host, entry.IP, false, 0, s.cfg.LeaseDuration())
		} else {
			log.Println(">> No host name")
		}
	} else {
		log.Println(">> No domain name")
	}
}

func (s *Service) getOptionsFromMAC(entry *MACEntry) dhcp4.Options {
	options := dhcp4.Options{}
	defaultOptions := s.cfg.Options()

	for i := range defaultOptions {
		options[i] = defaultOptions[i]
		log.Printf("OPTION:[%d][%+v]\n", i, defaultOptions[i])
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
			if domain := s.cfg.Domain(); domain != "" {
				options[dhcp4.OptionDomainName] = []byte(domain)
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
