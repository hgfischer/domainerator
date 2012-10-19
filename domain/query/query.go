package query

import (
	"fmt"
	"github.com/miekg/dns"
	"math/rand"
	"os"
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
	return fmt.Sprintf("%s\t%s\t%q\n", dr.Domain, dns.Rcode_str[dr.Rcode], dr.err)
}

// Return true if the domain is available (DNS NXDOMAIN)
func (dr Result) Available() bool {
	return dr.Rcode == dns.RcodeNameError
}

// Returns true if domain has a Name Server associated
func queryNS(domain string, dnsServers []string) (int, error) {
	c := new(dns.Client)
	c.ReadTimeout = time.Duration(4 * time.Second)
	c.WriteTimeout = time.Duration(4 * time.Second)
	c.Net = "udp"
	c.Retry = true
	c.Attempts = 3
	m := new(dns.Msg)
	m.RecursionDesired = true
	var err error
	for i := 0; i < 4; i++ {
		dnsServer := dnsServers[rand.Intn(len(dnsServers))]
		m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
		in, err := c.Exchange(m, dnsServer+":53")
		if err == nil {
			return in.Rcode, nil
		}
		time.Sleep(time.Duration(1 * time.Second))
	}
	return dns.RcodeRefused, err
}

// Check if each domain 
func CheckDomains(in chan string, out chan Result, dnsServers []string) {
	for domain := range in {
		var rCode int
		var err error
		rCode, err = queryNS(domain, dnsServers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to check domain %q: %q!\n", domain, err)
		}
		out <- Result{domain, rCode, err}
	}
}
