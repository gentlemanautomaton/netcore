package netdns

// FIXME: Restore WOL functionality
/*
func wakeByMAC(cfg *Config, mac net.HardwareAddr) error {
	return wol.SendMagicPacket(mac.String(), "255.255.255.255:9", "")
}

func wakeByIP(cfg *Config, ip net.IP) error {
	entry, err := cfg.db.GetIP(ip)
	if err != nil {
		return err
	}
	return wakeByMAC(cfg, entry.MAC)
}

func wakeByHostname(cfg *Config, hostname string) error {
	entry, err := cfg.db.GetDNS(hostname, "A")
	if err == nil {
		for i := range entry.Values {
			ip := net.ParseIP(entry.Values[i].Value)
			if ip != nil {
				err = wakeByIP(cfg, ip) // FIXME: Make
			}
			// FIXME: Find some better way of handling errors here?
		}
	}
	return err
}
*/
