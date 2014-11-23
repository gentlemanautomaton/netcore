package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

func dnsSetup(cfg *Config, etc *etcd.Client) chan error {
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) { dnsQueryServe(cfg, etc, w, req) })
	etc.CreateDir("dns", 0)
	exit := make(chan error, 1)

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "tcp", nil) // TODO: should use cfg to define the listening ip/port
	}()

	go func() {
		exit <- dns.ListenAndServe("0.0.0.0:53", "udp", nil) // TODO: should use cfg to define the listening ip/port
	}()

	return exit
}

func dnsQueryServe(cfg *Config, etc *etcd.Client, w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		fmt.Printf("DNS Query IS BOGUS %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())
		return
	}

	fmt.Printf("DNS Query %s %s from %s.\n", q.Name, dns.Type(q.Qtype).String(), w.RemoteAddr())

	// TODO: lookup in an in-memory cache (obeying TTLs!)

	qType := dns.Type(q.Qtype).String()                              // query type
	pathParts := strings.Split(strings.TrimSuffix(q.Name, "."), ".") // breakup the queryed name
	queryPath := strings.Join(reverseSlice(pathParts), "/")          // reverse and join them with a slash delimiter
	key := strings.ToLower("/dns/" + queryPath + "/@" + qType)       // structure the lookup key
	response, err := etc.Get(strings.ToLower(key), true, true)       // do the lookup
	if err == nil && response != nil && response.Node != nil && len(response.Node.Nodes) > 0 {
		var vals *etcd.Node
		meta := make(map[string]string)
		for _, node := range response.Node.Nodes {
			nodeKey := strings.Replace(node.Key, key+"/", "", 1)
			if nodeKey == "val" && node.Dir {
				vals = node
			} else if !node.Dir {
				meta[nodeKey] = node.Value // NOTE: the keys are case-sensitive
			}
		}

		answerMsg := new(dns.Msg)
		answerMsg.Id = req.Id
		answerMsg.Response = true
		answerMsg.Authoritative = true
		answerMsg.Question = req.Question
		answerMsg.Rcode = dns.RcodeSuccess
		answerMsg.Extra = []dns.RR{}
		answerTTL := uint32(10800) // this is the default TTL = 3 hours
		gotTTL, _ := strconv.Atoi(meta["TTL"])
		if gotTTL > 0 {
			answerTTL = uint32(gotTTL)
		}

		switch qType {
		case "SOA":
			answer := new(dns.SOA)
			answer.Header().Name = q.Name
			answer.Header().Ttl = answerTTL
			answer.Header().Rrtype = dns.TypeSOA
			answer.Header().Class = dns.ClassINET
			answer.Ns = strings.TrimSuffix(meta["NS"], ".") + "."
			answer.Mbox = strings.TrimSuffix(meta["MBOX"], ".") + "."
			answer.Serial = uint32(time.Now().Unix())
			answer.Refresh = uint32(60) // only used for master->slave timing
			answer.Retry = uint32(60)   // only used for master->slave timing
			answer.Expire = uint32(60)  // only used for master->slave timing
			answer.Minttl = uint32(60)  // how long caching resolvers should cache a miss (NXDOMAIN status)
			answerMsg.Answer = append(answerMsg.Answer, answer)
		default:
			// ... for answers that have values
			if vals != nil && vals.Nodes != nil {
				for _, child := range vals.Nodes {
					if child.Expiration != nil && child.Expiration.Unix() < time.Now().Unix() {
						continue
					}
					if child.TTL > 0 && uint32(child.TTL) < answerTTL {
						answerTTL = uint32(child.TTL)
					}

					// this builds the attributes for complex types, like MX and SRV
					attr := make(map[string]string)
					if child.Nodes != nil {
						for _, attrNode := range child.Nodes {
							nodeKey := strings.Replace(attrNode.Key, child.Key+"/", "", 1)
							attr[nodeKey] = attrNode.Value
						}
					}

					switch qType {
					// FIXME: Add more RR types!
					//        http://godoc.org/github.com/miekg/dns has info as well as
					//        http://en.wikipedia.org/wiki/List_of_DNS_record_types
					case "TXT":
						answer := new(dns.TXT)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeTXT
						answer.Header().Class = dns.ClassINET
						answer.Txt = []string{child.Value}
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "A":
						answer := new(dns.A)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeA
						answer.Header().Class = dns.ClassINET
						answer.A = net.ParseIP(child.Value)
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "AAAA":
						answer := new(dns.AAAA)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeAAAA
						answer.Header().Class = dns.ClassINET
						answer.AAAA = net.ParseIP(child.Value)
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "NS":
						answer := new(dns.NS)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeNS
						answer.Header().Class = dns.ClassINET
						answer.Ns = strings.TrimSuffix(child.Value, ".") + "."
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "CNAME":
						// FIXME: This is not being used quite correctly.  For example, we
						//        we need to ensure that we are not supplying conflicting
						//        RRs in the same space.  We need to be using the CNAME if
						//        it exists even if the client is asking for a different
						//        RR type.  We have a CNAME in this space and they ask for
						//        an A record?  Return the CNAME!  And maybe even resolve
						//        the value for them.  Or at least carry them as far down
						//        the CNAME chain as possible depending on our willingness
						//        to talk to external DNS servers for recursive queries.
						//        For now, I consider the present behavior incredibly
						//        broken.  It needs to be fixed.  See also DNAME.
						//				... http://en.wikipedia.org/wiki/CNAME_record
						answer := new(dns.CNAME)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeCNAME
						answer.Header().Class = dns.ClassINET
						answer.Target = strings.TrimSuffix(child.Value, ".") + "."
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "DNAME":
						// FIXME: This is not being used correctly.  See the notes about
						//        fixing CNAME and then consider that DNAME takes it a
						//        big step forward and aliases an entire subtree, not just
						//        a single name in the tree.  Note that this is for pointing
						//        to subtree, not to the self-equivalent.  See the Wikipedia
						//        article about it, linked below.  See also CNAME above.
						//				... http://en.wikipedia.org/wiki/CNAME_record#DNAME_record
						answer := new(dns.DNAME)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeDNAME
						answer.Header().Class = dns.ClassINET
						answer.Target = strings.TrimSuffix(child.Value, ".") + "."
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "PTR":
						answer := new(dns.PTR)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypePTR
						answer.Header().Class = dns.ClassINET
						answer.Ptr = strings.TrimSuffix(child.Value, ".") + "."
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "MX":
						answer := new(dns.MX)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeMX
						answer.Header().Class = dns.ClassINET
						answer.Preference = 50 // default if not defined
						priority, err := strconv.Atoi(attr["PRIORITY"])
						if err == nil {
							answer.Preference = uint16(priority)
						}
						if target, ok := attr["TARGET"]; ok {
							answer.Mx = strings.TrimSuffix(target, ".") + "."
						} else if child.Value != "" { // allows for simplified setting
							answer.Mx = strings.TrimSuffix(child.Value, ".") + "."
						}
						// FIXME: are we supposed to be returning these in prio ordering?
						//        ... or maybe it does that for us?  or maybe it's the enduser's problem?
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "SRV":
						answer := new(dns.SRV)
						answer.Header().Name = q.Name
						answer.Header().Rrtype = dns.TypeSRV
						answer.Header().Class = dns.ClassINET
						answer.Priority = 50 // default if not defined
						priority, err := strconv.Atoi(attr["PRIORITY"])
						if err == nil {
							answer.Priority = uint16(priority)
						}
						answer.Weight = 50 // default if not defined
						weight, err := strconv.Atoi(attr["WEIGHT"])
						if err == nil {
							answer.Weight = uint16(weight)
						}
						answer.Port = 50 // default if not defined
						port, err := strconv.Atoi(attr["PORT"])
						if err == nil {
							answer.Port = uint16(port)
						}
						if target, ok := attr["TARGET"]; ok {
							answer.Target = strings.TrimSuffix(target, ".") + "."
						} else if child.Value != "" { // allows for simplified setting
							targetParts := strings.Split(child.Value, ":")
							answer.Target = strings.TrimSuffix(targetParts[0], ".") + "."
							port, err := strconv.Atoi(targetParts[1])
							if err == nil {
								answer.Port = uint16(port)
							}
						}
						// FIXME: are we supposed to be returning these rando-weighted and in priority ordering?
						//        ... or maybe it does that for us?  or maybe it's the enduser's problem?
						answerMsg.Answer = append(answerMsg.Answer, answer)
					case "SSHFP":
						// TODO: implement SSHFP
						//       http://godoc.org/github.com/miekg/dns#SSHFP
					}
				}
			}
		}

		if len(answerMsg.Answer) > 0 {
			for _, answer := range answerMsg.Answer {
				answer.Header().Ttl = answerTTL
			}
			fmt.Printf("OUR DATA: [%+v]\n", answerMsg)
			w.WriteMsg(answerMsg)

			// TODO: cache the response locally in RAM?

			return
		}
	}

	// TODO: check to see if we host this zone; if yes, return NXDOMAIN right now!

	if false { // XXX: just for devtime!
		forwarders := cfg.DNSForwarders()
		if len(forwarders) == 0 {
			// we have no upstreams, so we'll just not use any
		} else if strings.TrimSpace(forwarders[0]) == "!" {
			// we've been told explicitly to not pass anything along to any upsteams
		} else {
			c := new(dns.Client)
			for _, server := range forwarders {
				c.Net = "udp"
				m, _, err := c.Exchange(req, strings.TrimSpace(server))

				if m != nil && m.MsgHdr.Truncated {
					c.Net = "tcp"
					m, _, err = c.Exchange(req, strings.TrimSpace(server))
				}

				if err != nil {
					fmt.Println(err)
				} else {
					w.WriteMsg(m)
					return // because we're done here
				}
			}
		}
	}

	// if we got here, it means we didn't find what we were looking for.
	failMsg := new(dns.Msg)
	failMsg.Id = req.Id
	failMsg.Response = true
	failMsg.Authoritative = true
	failMsg.Question = req.Question
	failMsg.Rcode = dns.RcodeNameError
	w.WriteMsg(failMsg)
	return
}

// FIXME: please support DNSSEC, verification, signing, etc...
