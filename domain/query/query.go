package query

import (
	"fmt"
	"github.com/miekg/dns"
	"math/rand"
	"time"
)

type Result struct {
	Domain string
	Rcode  int
	err    error
}

// Format Result into string for output file
func (dr Result) String(simple bool) string {
	if simple {
		return fmt.Sprintf("%s\n", dr.Domain)
	}
	return fmt.Sprintf("%s\t%s\t%q\t\n", dr.Domain, dns.Rcode_str[dr.Rcode], dr.err)
}

// Return true if the domain is available (DNS NXDOMAIN)
func (dr Result) Available() bool {
	return dr.Rcode == dns.RcodeNameError
}

// Returns true if domain has a Name Server associated
func queryNS(domain string, dnsServers []string, proto string) (int, error) {
	c := new(dns.Client)
	c.ReadTimeout = time.Duration(2 * time.Second)
	c.WriteTimeout = time.Duration(2 * time.Second)
	c.Net = proto
	m := new(dns.Msg)
	m.RecursionDesired = true
	dnsServer := dnsServers[rand.Intn(len(dnsServers))]
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	in, err := c.Exchange(m, dnsServer+":53")
	if err == nil {
		return in.Rcode, err
	}
	return dns.RcodeRefused, err
}

// Check if each domain 
func CheckDomains(id int, in, retries chan string, out chan Result, dnsServers []string, proto string) {
	for {
		var domain string
		select {
		case domain = <-in:
		case domain = <-retries:
		}
		rCode, err := queryNS(domain, dnsServers, proto)
		if err != nil {
			retries <- domain
		} else {
			out <- Result{domain, rCode, err}
		}
	}
}
