package rdap

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	// RDAP Server URL's
	// From https://about.rdap.org/
	DefaultServer = "https://rdap.org"
	CentralNic    = "https://www.centralnic.com/registry/labs/rdap"
	VerisignDNRD  = "http://dnrd.verisignlabs.com/dnrd-ap/help/"
	APNIC         = "https://www.apnic.net/apnic-info/whois_search/about/rdap"
	ARIN          = "https://www.arin.net/resources/whoisrws/"
	LACNIC        = "http://restfulwhoisv2.labs.lacnic.net/restfulwhois/"
	RIPE          = "https://rdap.db.ripe.net"

	// Setup for the default http.Client used
	DefaultHTTPClient = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,

				InsecureSkipVerify: true, // TODO(adam)
			},
			TLSHandshakeTimeout:   1 * time.Minute,
			IdleConnTimeout:       1 * time.Minute,
			ResponseHeaderTimeout: 1 * time.Minute,
			ExpectContinueTimeout: 1 * time.Minute,
		},
		Timeout: 30 * time.Second,
	}

	// RFC7480 Section 4.2 and RFC7483 Section 10.1
	// mention either application/json or application/rdap+json
	// can be passed on the Accept header
	//
	// However, some implementations don't support the new content type
	// so we use the standard json mime type.
	DefaultAcceptHeader = "application/json"
)

// Client is is used to make RDAP HTTP requests against a server.
type Client struct {
	Underlying *http.Client

	// Which RDAP server address to use, by default will be the
	// RDAP official server.
	BaseAddress string

	setup sync.Once
}

// IP represents a /ip/$foo request, where $foo is either an IPv4, IPv6
// address or a CIDR network range.
func (c *Client) IP(addr string) (*IPNetwork, error) {
	ip := net.ParseIP(addr)
	_, net, _ := net.ParseCIDR(addr)

	// sanity check the input
	if ip == nil && net == nil {
		return nil, fmt.Errorf("invalid ip or cidr specified: %q", addr)
	}
	if net != nil && (len(net.IP) == 0 && len(net.Mask) == 0) {
		return nil, fmt.Errorf("invalid ip or cidr specified: %q", addr)
	}

	// set parsed form back as input
	if len(ip) > 0 {
		addr = ip.String()
	}
	if net != nil && len(net.IP) > 0 {
		addr = net.String()
	}

	// Build and make the actual HTTP request
	req, err := c.makeRequest(fmt.Sprintf("/ip/%s", addr))
	fmt.Println(req.URL)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("no body on successful response for %s", req.URL)
	}
	defer resp.Body.Close()

	// Parse successful response
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body from %s", req.URL)
	}
	fmt.Println(string(bs))
	var ipNetwork IPNetworkJSON
	if err := json.Unmarshal(bs, &ipNetwork); err != nil {
		return nil, fmt.Errorf("error parsing ip network response: %v", err)
	}
	fmt.Println("B: ", ipNetwork)
	return nil, errors.New("") // TODO(Adam): parse successful response
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
func (c *Client) Domain(fqdn string) (*Domain, error) {
	// TODO(adam): parse domain?
	if fqdn == "" {
		return nil, errors.New("empty FQDN provided")
	}

	req, err := c.makeRequest(fmt.Sprintf("/domain/%s", fqdn))
	fmt.Println(req.URL)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("no body on successful response for %s", req.URL)
	}
	defer resp.Body.Close()

	// Parse successful response
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body from %s", req.URL)
	}
	var domain Domain
	if err := json.Unmarshal(bs, &domain); err != nil {
		return nil, fmt.Errorf("error parsing domain response: %v", err)
	}
	if domain.ObjectClassName != "domain" {
		return &domain, fmt.Errorf("unknown objectClassName: %q", domain.ObjectClassName)
	}
	return &domain, nil
}

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

func (c *Client) makeRequest(seg string) (*http.Request, error) {
	u, err := url.Parse(strings.TrimSuffix(c.BaseAddress, "/"))
	if err != nil {
		return nil, err
	}
	return http.NewRequest("GET", fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path.Join(u.Path, seg)), nil)
}

// do is a helper method which will initialize some internal properties of a
// Client (if not already set) and perform some sanity checks on the request.
//
// If the underlying HTTP call fails do will attempt to read out an Error message
// and close the response body.
//
// On a successful request do will not close or alter the response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	c.setup.Do(func() {
		if c.Underlying == nil {
			c.Underlying = DefaultHTTPClient
		}

		if c.BaseAddress == "" {
			c.BaseAddress = DefaultServer
		} else {
			// Drop the trailing slash so we don't accidently double up
			// in url building.
			c.BaseAddress = strings.TrimSuffix(c.BaseAddress, "/")
		}
		if req.URL.Host == "" {
			raw := DefaultServer + req.URL.Path
			u, err := url.Parse(raw)
			if err != nil {
				panic(fmt.Errorf("invalid url %q: %v", raw, err))
			}
			req.URL = u
		}
	})

	// RFC7481 Section 3.5
	// As noted in Section 3.2, the HTTP "basic" authentication scheme can
	// be used to authenticate a client.  When this scheme is used, HTTP
	// over TLS MUST be used to protect the client's credentials from
	// disclosure while in transit.
	v := req.Header.Get("Authentication")
	if v != "" && req.URL.Scheme != "https" {
		return nil, fmt.Errorf("invalid scheme %q in request with Authentication header", req.URL.Scheme)
	}

	// Set Accept header if it's not already added
	if v := req.Header.Get("Accept"); v == "" {
		req.Header.Set("Accept", DefaultAcceptHeader)
	}

	// Perform the request
	resp, err := c.Underlying.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during request: %v", err)
	}

	// TODO(Adam): We should check for both of these
	// Clients must follow redirects // RFC7480 Section 5.2
	// 429 status, return Retry-After header // RFC7480 Section 5.5

	if resp.StatusCode >= 400 {
		if resp.Body != nil {
			// RFC7480 Section 5.3 states servers MAY return an error response
			// so we will try and parse that out from the body
			return nil, c.parseError(resp.Body)
		}
		// RFC7480 Sectin 5.3:
		// If a server wishes to inform the client that information about the
		// query is available, but cannot include the information in the
		// response to the client for policy reasons, the server MUST respond
		// with an appropriate response code out of HTTP's 4xx range.
		return nil, fmt.Errorf("%d error during request to %s", resp.StatusCode, req.URL)
	}
	return resp, nil
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
		return nil
	}
	return &err
}
