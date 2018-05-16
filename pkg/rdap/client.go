package rdap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
)

var (
	// RDAP Server URL's
	// From https://about.rdap.org/
	DefaultServer = "https://rdap.org"
	CentralNic = "https://www.centralnic.com/registry/labs/rdap"
	VerisignDNRD = "http://dnrd.verisignlabs.com/dnrd-ap/help/"
	APNIC = "https://www.apnic.net/apnic-info/whois_search/about/rdap"
	ARIN = "https://www.arin.net/resources/whoisrws/"
	LACNIC = "http://restfulwhoisv2.labs.lacnic.net/restfulwhois/"
	RIPE = "https://rdap.db.ripe.net"
)

// Client is an RDAP compatiable
//
// RFC7482 Section 3.1
// o  'ip': Used to identify IP networks and associated data referenced
//    using either an IPv4 or IPv6 address.

// o  'autnum': Used to identify Autonomous System number registrations
//    and associated data referenced using an asplain Autonomous System
//    number.

// o  'domain': Used to identify reverse DNS (RIR) or domain name (DNR)
//    information and associated data referenced using a fully qualified
//    domain name.

// o  'nameserver': Used to identify a nameserver information query
//    using a host name.

// o  'entity': Used to identify an entity information query using a
//    string identifier.

type Client struct {
	Underlying *http.Client

	// TODO(adam)
	BaseAddress string

	setup sync.Once
}

// RFC7482 3.1.1.  IP Network Path Segment Specification
//    Syntax: ip/<IP address> or ip/<CIDR prefix>/<CIDR length>
//
// IPv4 dotted decimal or IPv6 [RFC5952] address OR
// an IPv4 or IPv6 Classless Inter-domain Routing (CIDR) [RFC4632] notation address block (i.e., XXX/YY)
func (c *Client) IP(addr string) (*IPNetwork, error) {
	ip := net.ParseIP(addr)
	_, net, _ := net.ParseCIDR(addr)

	// sanity check the input
	if ip == nil && net == nil {
		return nil, fmt.Errorf("invalid ip or cidr specified: %q", addr)
	}
	if len(ip) == 0 && (len(net.IP) == 0 || len(net.Mask) == 0) {
		return nil, fmt.Errorf("invalid ip or cidr specified: %q", addr)
	}

	// set parsed form back as input
	if len(ip) > 0 {
		addr = ip.String()
	}
	if len(net.IP) > 0 {
		addr = net.String()
	}

	// Build and make the actual HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/ip/%s", c.BaseAddress, addr), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.do(req)
	if err != nil {
		return nil, err
	}
	return nil, errors.New("") // TODO(Adam)
}

// RFC7482 3.1.2.  Autonomous System Path Segment Specification
//    Syntax: autnum/<autonomous system number>
//
// /autnum/XXX/ ... where XXX is an asplain Autonomous System number [RFC5396]
// TODO(adam): Does RFC5396 specify any format?
func (c *Client) Autnum() {}

// RFC7482 3.1.3.  Domain Path Segment Specification
//    Syntax: domain/<domain name>
//
// Queries for domain information are of the form /domain/XXXX/...,
// where XXXX is a fully qualified (relative to the root) domain name
// (as specified in [RFC0952] and [RFC1123]) in either the in-addr.arpa
// or ip6.arpa zones (for RIRs) or a fully qualified domain name in a
// zone administered by the server operator (for DNRs).
//
// Internationalized Domain Names (IDNs) represented in either A-label
// or U-label format [RFC5890] are also valid domain names.
func (c *Client) Domain() {}

// RFC7482 3.2.1 Domain Search
//
// /domains?name=XXXX
// XXXX is a search pattern representing a domain name in "letters,
// digits, hyphen" (LDH) format [RFC5890] in a zone administered by the
// server operator of a DNR.
//
// Searches for domain information by nameserver name are specified
// using this form:
// /domains?nsLdhName=YYYY
// YYYY is a search pattern representing a host name in "letters,
// digits, hyphen" format [RFC5890] in a zone administered by the server
// operator of a DNR.
//
// Searches for domain information by nameserver IP address are
// specified using this form:
// /domains?nsIp=ZZZZ
// ZZZZ is a search pattern representing an IPv4 [RFC1166] or IPv6
// [RFC5952] address.
//
// RFC7483 Section 6
// for /domains searches, the array is "domainSearchResults"
func (c *Client) DomainSearch() {}

// RFC7482 3.1.4.  Nameserver Path Segment Specification
//    Syntax: nameserver/<nameserver name>
//
// The <nameserver name> parameter represents a fully qualified host
// name as specified in [RFC0952] and [RFC1123].  Internationalized
// names represented in either A-label or U-label format [RFC5890] are
// also valid nameserver names.
func (c *Client) Nameserver() {}

// RFC7482 3.2.2.  Nameserver Search
// Syntax: nameservers?name=<nameserver search pattern>
// Syntax: nameservers?ip=<nameserver search pattern>
//
// RFC7483 Section 6
// for /nameservers searches, the array is "nameserverSearchResults"
func (c *Client) NameserverSearch() {}

// RFC7482 3.1.5.  Entity Path Segment Specification
//    Syntax: entity/<handle>
//
// The <handle> parameter represents an entity (such as a contact,
// registrant, or registrar) identifier whose syntax is specific to the
// registration provider.  For example, for some DNRs, contact
// identifiers are specified in [RFC5730] and [RFC5733].
func (c *Client) Entity() {}

// RFC7482 3.2.3.  Entity Search
// Syntax: entities?fn=<entity name search pattern>
// Syntax: entities?handle=<entity handle search pattern>
//
// entities?fn=XXXX
// XXXX is a search pattern representing the "FN" property of an entity
// (such as a contact, registrant, or registrar) name as specified in
// Section 5.1 of [RFC7483].
//
// entities?handle=XXXX
// XXXX is a search pattern representing an entity (such as a contact,
// registrant, or registrar) identifier whose syntax is specific to the
// registration provider.
//
// RFC7483 Section 6
// for /entities searches, the array is "entitySearchResults"
func (c *Client) EntitySearch() {}

// RFC7482 3.1.6.  Help Path Segment Specification
//    Syntax: help
// The help path segment can be used to request helpful information
// (command syntax, terms of service, privacy policy, rate-limiting
// policy, supported authentication methods, supported extensions,
// technical support contact, etc.) from an RDAP server.
func (c *Client) Help() {}
// RFC7483 Section 7
// The appropriate response to /help queries as defined by [RFC7482] is
// to use the notices structure as defined in Section 4.3.


// Clients must follow redirects // RFC7480 Section 5.2
// error on 404, try and parse resp.Body still // RFC7480 Section 5.3
// 429 status, return Retry-After header // RFC7480 Section 5.5
func (c *Client) do(req *http.Request) (*http.Response, error) {
	c.setup.Do(func(){
		if c.Underlying == nil {
			c.Underlying = &http.Client{} // TODO(adam): set default in this package
		}

		if c.BaseAddress == "" {
			c.BaseAddress = DefaultServer
		}
		if req.URL.Host == "" {
			u, err := url.Parse(DefaultServer + req.URL.Path)
			if err != nil {
				panic(err) // TODO
			}
			req.URL = u
		}
	})

	resp, err := c.Underlying.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s", err) // TODO(Adam):
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 499 {
		// try and marshal out an Error
		if resp.Body != nil {
			return nil, c.parseError(resp.Body)
		}
		return nil, nil // TODO
	}
	return resp, nil

	// RFC7481 Section 3.5
	// As noted in Section 3.2, the HTTP "basic" authentication scheme can
	// be used to authenticate a client.  When this scheme is used, HTTP
	// over TLS MUST be used to protect the client's credentials from
	// disclosure while in transit.
}

// parseError attempts to parse `bs` as an Error type
// a nil response means no error was parsed
//
// The reader given to parseError will be closed
func (c *Client) parseError(r io.ReadCloser) *Error {
	defer r.Close()

	bs, e := ioutil.ReadAll(r)
	if e != nil {
		return nil
	}
	var err Error
	if e := json.Unmarshal(bs, &err); e != nil {
		fmt.Println(e)
		return nil
	}
	return &err
}
