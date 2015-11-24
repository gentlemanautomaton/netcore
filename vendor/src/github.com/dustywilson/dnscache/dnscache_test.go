package dnscache

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// TODO: Write some tests

type samplePool struct {
	queries []dns.Question
	answers map[dns.Question][]dns.RR
}

type generator struct {
	rand.Source
}

func newGenerator(seed int64) *generator {
	return &generator{rand.NewSource(seed)}
}

func (g *generator) RandString(n int) string {
	// See http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, g.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = g.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func (g *generator) RandFQDN(i int) string {
	output := g.RandString(i % 23)
	if i%3 == 0 {
		output += "." + g.RandString(i%19)
	}
	if i%7 == 0 {
		output += "." + g.RandString(i%29)
	}
	return output
}

func (g *generator) DeterministicIPv4(i int) net.IP {
	a := byte((i + 31) % 255)
	b := byte((i + 67) % 255)
	c := byte((i + 97) % 255)
	d := byte((i + 157) % 255)
	return net.IPv4(a, b, c, d)
}

func generateSamplePool(size int, seed int64) *samplePool {
	g := newGenerator(seed)
	sample := &samplePool{
		queries: make([]dns.Question, 0, size),
		answers: make(map[dns.Question][]dns.RR),
	}
	for i := 0; i < size; i++ {
		var qtype uint16
		switch i % 2 {
		case 0:
			qtype = dns.TypeTXT
		case 1:
			qtype = dns.TypeA
		}
		question := dns.Question{
			Name:   g.RandFQDN(i),
			Qtype:  qtype,
			Qclass: dns.ClassINET,
		}
		sample.queries = append(sample.queries, question)
		var answer dns.RR
		switch question.Qtype {
		case dns.TypeTXT:
			answer = answerTXT(question, g.RandString(i%31))
		case dns.TypeA:
			answer = answerA(question, g.DeterministicIPv4(i))
		}
		sample.answers[question] = append([]dns.RR(nil), answer)
	}
	return sample
}

func (sp *samplePool) LookupFunc() func(dns.Question) []dns.RR {
	return func(q dns.Question) []dns.RR {
		return sp.answers[q]
	}
}

func answerTXT(q dns.Question, value string) dns.RR {
	answer := new(dns.TXT)
	answer.Header().Name = q.Name
	answer.Header().Rrtype = dns.TypeTXT
	answer.Header().Class = dns.ClassINET
	answer.Txt = []string{value}
	return answer
}

func answerA(q dns.Question, value net.IP) dns.RR {
	answer := new(dns.A)
	answer.Header().Name = q.Name
	answer.Header().Rrtype = dns.TypeA
	answer.Header().Class = dns.ClassINET
	answer.A = value
	return answer
}

func newSampleCache(poolSize, bufferSize int, seed int64, maxTTL, missTTL time.Duration) (*samplePool, *Cache) {
	sample := generateSamplePool(poolSize, seed)
	cache := New(bufferSize, maxTTL, missTTL, sample.LookupFunc())
	return sample, cache
}

func printRecords(rr []dns.RR) {
	if len(rr) == 1 {
		fmt.Printf("%v\n", rr[0])
	} else {
		fmt.Printf("%v\n", rr)
	}
}

var tempRR []dns.RR

func benchmark(b *testing.B, poolSize, bufferSize int, seed int64) {
	sample, cache := newSampleCache(poolSize, bufferSize, seed, time.Minute*5, time.Second*30)
	//total := b.N
	total := b.N
	requests := make([]Request, total)
	k := 0
	for k < total {
		for _, q := range sample.queries {
			requests[k] = Request{
				Question:     q,
				ResponseChan: make(chan []dns.RR),
			}
			k++
			if k >= total {
				break
			}
		}
	}
	b.ResetTimer()
	for i := range requests {
		cache.Lookup(requests[i])
	}
	for i := range requests {
		tempRR = <-requests[i].ResponseChan
	}
	cache.Stop()
}

func BenchmarkRandPool100(b *testing.B)  { benchmark(b, 100, 0, 78897) }
func BenchmarkRandPool1000(b *testing.B) { benchmark(b, 1000, 0, 215487113) }
