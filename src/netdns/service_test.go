package netdns

import (
	"testing"

	"github.com/miekg/dns"
)

func TestWOL(t *testing.T) {
	// TODO: Create database mockup that we can query?
	m := new(dns.Msg)
	m.SetQuestion("_wol.test", dns.TypeTXT)
}
