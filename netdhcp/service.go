package netdhcp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sort"
	"sync"

	"context"

	"github.com/krolaw/dhcp4"
)

// Service provides netcore DHCP services.
type Service struct {
	mutex   sync.RWMutex // guards access to v
	p       Provider
	id      string
	v       *service // current view of the service, protected by mutex
	started chan bool
	done    chan Completion
}

// NewService creates a new netcore DHCP service with the given data provider
// service instance ID.
func NewService(provider Provider, instance string) *Service {
	s := &Service{
		p:       provider,
		id:      instance,
		started: make(chan bool, 1),
		done:    make(chan Completion, 1),
	}

	go s.init()

	return s
}

func (s *Service) init() {
	s.mutex.RLock()
	id := s.id
	p := s.p
	s.mutex.RUnlock()

	if p == nil {
		s.failedInit(errors.New("No provider given."))
		return
	}

	if id == "" {
		s.failedInit(errors.New("No instance identifier given."))
		return
	}

	view, err := s.build(p, id)
	if err != nil {
		s.failedInit(err)
		return
	}

	s.signalStarted(true)
	err = dhcp4.ListenAndServeIf(view.config.ServerNIC(), s)
	s.signalDone(true, err)
}

func (s *Service) failedInit(err error) {
	s.signalStarted(false)
	s.signalDone(false, err)
}

func (s *Service) signalStarted(success bool) {
	s.started <- success
	close(s.started)
}

func (s *Service) signalDone(initialized bool, err error) {
	s.done <- Completion{false, err}
	close(s.done)
}

func (s *Service) build(p Provider, id string) (*service, error) {
	ctx := context.Background() // FIXME: Use a real cancellable context

	gc := NewContext(p)
	global, err := gc.Read(ctx)
	if err != nil {
		return nil, err
	}

	ic := gc.Instance(id)
	instance, err := ic.Read(ctx)
	if err != nil {
		return nil, err
	}

	nid := MergeNetworkID(global, instance)
	if nid == "" {
		return nil, ErrNoConfigNetwork
	}

	nc := gc.Network(nid)
	network, err := nc.Read(ctx)
	if err != nil {
		return nil, err
	}

	config := MergeServerConfig(global, instance, network) // TODO: Review order
	if err := ValidateServerConfig(&config); err != nil {
		return nil, err
	}

	attr := MergeBindingConfig(global, instance, network) // TODO: Review order
	if err := ValidateBindingConfig(&attr); err != nil {
		return nil, err
	}

	srv := &service{
		id:       id,
		global:   gc,
		instance: ic,
		network:  nc,
		config:   config,
		attr:     attr,
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.v = srv

	return srv, nil
}

// view returns a consistent view of the service.
func (s *Service) view() *service {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.v
}

type params struct {
	p  Provider
	id string
}

/*
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
*/

func (s *Service) sync(chan params) {
	// TODO: Figure this out
	/*
		var (
			gs       <-chan GlobalUpdate
			is       <-chan InstanceUpdate
			ns       <-chan NetworkUpdate
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
	*/
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
	return s.view().ServeDHCP(packet, msgType, reqOptions)
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
	config   Config
	attr     Attr
	//cache    cache
}

func (s *service) ServeDHCP(packet dhcp4.Packet, msgType dhcp4.MessageType, reqOptions dhcp4.Options) (response dhcp4.Packet) {
	ctx := context.TODO() // FIXME: Use a real cancellable context

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

		/*
			data, found, err := s.net.MAC.Lookup(context.Background(), addr)
			if err != nil {
				// FIXME: Log error?
				return nil
			}
		*/

		binding, err := s.Binding(ctx, addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		_, _, err = s.selectBinding(context.Background(), &binding, addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		/*
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
			ip = s.getIPFromPool()
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
		*/

		//log.Printf("DHCP Discover from %s (no offer due to no addresses available in pool)\n", mac.String())
		// FIXME: Send to StatHat and/or increment a counter
		// TODO: Send an email?

		return nil

	case dhcp4.Request:
		// RFC 2131 4.3.2
		// FIXME: send to StatHat and/or increment a counter
		addr := packet.CHAddr()

		// Check MAC blacklist
		if !s.isMACPermitted(addr) {
			log.Printf("DHCP Request from %s\n is not permitted", addr.String())
			return nil
		}

		// Check IP presence
		state, requestedIP := s.getRequestState(packet, reqOptions)
		log.Printf("DHCP Request (%s) from %s...\n", state, addr.String())
		if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // no IP provided at all... why? FIXME
			log.Printf("DHCP Request (%s) from %s (empty IP, so we're just ignoring this request)\n", state, addr.String())
			return nil
		}

		// Check IPv4
		if len(requestedIP) != net.IPv4len {
			log.Printf("DHCP Request (%s) from %s wanting %s (IPv6 address requested, so we're just ignoring this request)\n", state, addr.String(), requestedIP.String())
			return nil
		}

		// Check IP subnet
		if !s.config.Subnet.Contains(requestedIP) {
			log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to wrong subnet)\n", state, addr.String(), requestedIP.String())
			return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.config.IP.To4(), nil, 0, nil)
		}

		// Check Target Server
		targetServerIP := packet.SIAddr()
		if len(targetServerIP) > 0 && !targetServerIP.IsUnspecified() {
			log.Printf("DHCP Request (%s) from %s wanting %s is in response to a DHCP offer from %s\n", state, addr.String(), requestedIP.String(), targetServerIP.String())
			if s.config.IP.Equal(targetServerIP) {
				return nil
			}
		}

		// Process Request
		log.Printf("DHCP Request (%s) from %s wanting %s...\n", state, addr.String(), requestedIP.String())

		binding, err := s.Binding(ctx, addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		_, _, err = s.selectBinding(context.Background(), &binding, addr)
		if err != nil {
			// FIXME: Log error?
			return nil
		}

		/*
			lease, found, err := s.p.MAC(mac, true)
			if err != nil {
				return nil
			}
		*/

		/*
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
		*/

		/*
			if err == nil {
				s.maintainDNSRecords(lease, packet, reqOptions) // TODO: Move this?
				options := s.getOptionsFromMAC(lease)
				log.Printf("DHCP Request (%s) from %s wanting %s (we agree)\n", state, mac.String(), requestedIP.String())
				return dhcp4.ReplyPacket(packet, dhcp4.ACK, s.cfg.IP().To4(), requestedIP.To4(), lease.Duration, options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
			}
		*/

		log.Printf("DHCP Request (%s) from %s wanting %s (we reject due to address collision)\n", state, addr.String(), requestedIP.String())
		return dhcp4.ReplyPacket(packet, dhcp4.NAK, s.config.IP.To4(), nil, 0, nil)

	case dhcp4.Decline:
		// RFC 2131 4.3.3
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		addr := packet.CHAddr()
		log.Printf("DHCP Decline from %s\n", addr.String())

	case dhcp4.Release:
		// RFC 2131 4.3.4
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		addr := packet.CHAddr()
		log.Printf("DHCP Release from %s\n", addr.String())

	case dhcp4.Inform:
		// RFC 2131 4.3.5
		// https://tools.ietf.org/html/draft-ietf-dhc-dhcpinform-clarify-06
		// FIXME: release from DB?  tick a flag?  increment a counter?  send to StatHat?
		// FIXME: we should reply with valuable info, but not assign an IP to this client, per RFC 2131 for DHCPINFORM
		// NOTE: the client's IP is supposed to only be in the ciaddr field, not the requested IP field, per RFC 2131 4.4.3
		//addr := packet.CHAddr()
		ip := packet.CIAddr()
		if len(ip) > 0 && !ip.IsUnspecified() {
			/*
				log.Printf("DHCP Inform from %s for %s \n", addr.String(), ip.String())
				if len(ip) == net.IPv4len && s.config.GuestPool().Contains(ip) {
					entry, found, _ := s.p.MAC(addr, true)
					if found {
						options := s.getOptionsFromMAC(entry)
						return informReplyPacket(packet, dhcp4.ACK, s.cfg.IP().To4(), options.SelectOrderOrAll(reqOptions[dhcp4.OptionParameterRequestList]))
					}
				}
			*/
		}
	}

	return nil
}

func (s *service) isMACPermitted(addr net.HardwareAddr) bool {
	// TODO: determine whether or not this MAC should be permitted to get an IP at all (blacklist? whitelist?)
	return true
}

func (s *service) pepareLease(target net.HardwareAddr) {

}

func (s *service) selectBinding(ctx context.Context, binding BindingConfig, target net.HardwareAddr) (*Assignment, *Lease, error) {
	// Enumerate all reservations and dynamic allocations and select the most
	// appropriate lease based on the following algorithm:
	// 1. Reservation with active lease matching this MAC, sorted by priority and then by recency
	// 2. Reservation without any active lease, sorted by priority and then by recency
	// 3. Dynamic Allocation with active lease matching this MAC and current pools, sorted by priority and then by recency
	// 4. Dynamic Allocation without any active lease matching the current pools, sorted by priority and then by recency
	// 5. New Dynamic Allocation from the current pools, sorted by priority

	// TODO: Consider whether we should always give the lease of highest priority
	//       regardless of whatever the current lease is.

	type result struct {
		Lease *Lease
		Err   error
	}

	//pools := binding.BindingPools()
	assignments := binding.BindingAssignments()

	// TODO: Use the pools

	if len(assignments) == 0 {
		// FIXME: Lack of assignments means we should select from the pool, not return an error
		return nil, nil, errors.New("No assignments specified.")
	}

	// TODO: Clone the assignment set?

	sort.Sort(assignments)

	// Issue a lease lookup for each potential IP address in parallel
	queries := make([]chan result, 0, len(assignments))
	for i := range assignments {
		// FIXME: Make sure the ip address is within this network
		ch := make(chan result)
		queries = append(queries, ch)
		go func(assignment *Assignment) {
			lease, err := s.network.Lease(assignment.Address).Read(ctx)
			ch <- result{lease, err}
		}(assignments[i])
	}

	// Process each IP and its resulting lease lookup in preferential order
	for i, ch := range queries {
		res := <-ch
		switch res.Err {
		case nil:
			if bytes.Equal(res.Lease.MAC, target) {
				// TODO: Renew the lease
				return assignments[i], res.Lease, nil
			}
		case ErrNotFound:
			// TODO: Attempt to create the lease
		default:
			// Something went horribly wrong
			return nil, nil, res.Err
		}
	}
	return nil, nil, nil
}

func (s *service) getRequestState(packet dhcp4.Packet, reqOptions dhcp4.Options) (string, net.IP) {
	state := "NEW"
	requestedIP := net.IP(reqOptions[dhcp4.OptionRequestedIPAddress])
	if len(requestedIP) == 0 || requestedIP.IsUnspecified() { // empty
		state = "RENEWAL"
		requestedIP = packet.CIAddr()
	}
	return state, requestedIP
}

/*
func (s *service) getLeaseDurationForRequest(reqOptions dhcp4.Options, defaultDuration time.Duration) time.Duration {
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

func (s *service) getIPFromPool() net.IP {
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
*/

/*
func (s *service) maintainDNSRecords(entry *MACEntry, packet dhcp4.Packet, reqOptions dhcp4.Options) {
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
			// FIXME: Make this work again
			//s.p.RegisterA(host, entry.IP, false, 0, s.cfg.LeaseDuration())
		} else {
			log.Println(">> No host name")
		}
	} else {
		log.Println(">> No domain name")
	}
}
*/

// Attr retrieves the DHCP lease configuration for the given MAC address by
// overlaying all of the possible configuration sources.
//
// Sources are overlayed in the following order:
// 1. Global
// 2. Network
// 3. Instance
// 4. Global MAC prefixes (least specific to most specific)
// 5. Network MAC prefixes (least specific to most specific)
// 6. Global type
// 7. Network type
// 8. Global device
// 9. Network device
//
// TODO: Maintain a cache of observers for recently retrieved MAC prefix data.
//       The observers should be closed and removed from the cache after some
//       duration since the last retrieval (which could be days).
func (s *service) Binding(ctx context.Context, addr net.HardwareAddr) (attr Attr, err error) {
	addrSet := macPrefixes(addr)
	gpSlice := make([]BindingConfig, 0, len(addrSet))
	npSlice := make([]BindingConfig, 0, len(addrSet))

	// Global MAC Prefixes
	for _, paddr := range macPrefixes(addr) {
		gprefix, perr := s.global.Prefix(paddr).Read(ctx)
		if perr == nil {
			gpSlice = append(gpSlice, gprefix)
		}
		nprefix, perr := s.network.Prefix(paddr).Read(ctx)
		if perr == nil {
			npSlice = append(npSlice, nprefix)
		}
	}
	gpAttr := MergeBindingConfig(gpSlice...)
	npAttr := MergeBindingConfig(npSlice...)
	attr = MergeBindingConfig(&s.attr, &gpAttr, &npAttr)
	return
}

/*
func (s *service) getOptionsFromMAC(entry *MACEntry) dhcp4.Options {
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
*/

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
