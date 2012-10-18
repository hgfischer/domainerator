package query

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"os"
	"time"
)

type Result struct {
	Domain string
	Rcode  int
}

// Format Result into string for output file
func (dr Result) String(simple bool) string {
	if simple {
		return fmt.Sprintf("%s\n", dr.Domain)
	}
	return fmt.Sprintf("%s\t%s\n", dr.Domain, dns.Rcode_str[dr.Rcode])
}

// Return true if the domain is available (DNS NXDOMAIN)
func (dr Result) Available() bool {
	return dr.Rcode == dns.RcodeNameError
}

// Returns true if domain has a Name Server associated
func queryNS(domain, dnsServer string) (int, error) {
	c := new(dns.Client)
	c.ReadTimeout = time.Duration(10 * time.Second)
	c.WriteTimeout = time.Duration(10 * time.Second)
	c.Net = "tcp"
	c.Retry = true
	c.Attempts = 5
	m := new(dns.Msg)
	m.RecursionDesired = true
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	in, err := c.Exchange(m, dnsServer+":53")
	if err != nil {
		return dns.RcodeRefused, err
	}
	return in.Rcode, err
}

// Check if each domain 
func CheckDomains(in chan string, out chan Result, dnsServer string) {
	for domain := range in {
		var rCode int
		var err error
		rCode, err = queryNS(domain, dnsServer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nFailed to check domain %q at DNS %q (%q)!\n", domain, dnsServer, err)
		}
		out <- Result{domain, rCode}
	}
}
