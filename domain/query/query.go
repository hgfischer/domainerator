package query

import (
	"fmt"
	"github.com/miekg/dns"
	"os"
	"time"
)

type Result struct {
	domain string
	rCode  int
}

// Format Result into string for output file
func (dr Result) String() string {
	return fmt.Sprintf("%s\t%s\n", dr.domain, dns.Rcode_str[dr.rCode])
}

// Returns true if domain has a Name Server associated
func queryNS(domain, dnsServer string) (int, error) {
	c := new(dns.Client)
	c.Net = "tcp"
	c.ReadTimeout = time.Duration(4) * time.Second
	c.WriteTimeout = c.ReadTimeout
	c.Retry = true
	c.Attempts = 3
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeNS)
	m.RecursionDesired = true
	in, err := c.Exchange(m, dnsServer+":53")
	if err != nil {
		return dns.RcodeNameError, err
	}
	return in.Rcode, err
}

// Check if each domain 
func CheckDomains(in chan string, out chan Result, dnsServer string) {
	for domain := range in {
		var rCode int
		var err error
		// try 5 times before failing
		for i := 0; i < 5; i++ {
			rCode, err = queryNS(domain, dnsServer)
			if err == nil {
				break
			}
		}
		if err == nil {
			out <- Result{domain, rCode}
		} else {
			fmt.Fprintf(os.Stderr, "\nFailed to check domain %q at DNS %q (%q)!\n", domain, dnsServer, err)
			out <- Result{domain, dns.RcodeServerFailure}
		}
	}
}
