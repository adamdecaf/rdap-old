package bootstrap

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	// From RDAP section on https://www.iana.org/protocols
	ASNUrl  = "https://data.iana.org/rdap/asn.json"
	DNSUrl  = "https://data.iana.org/rdap/dns.json"
	IPv4Url = "https://data.iana.org/rdap/ipv4.json"
	IPv6Url = "https://data.iana.org/rdap/ipv6.json"

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
)

// Registry is a type which returns the RDAP server for a given
// domain, ip network or AS number.
//
// The default implementation does not perform any caching and instead
// calls its Grabber on every request.
type Registry struct {
	// Endpoints
	ASNEndpoint  string
	DNSEndpoint  string
	IPv4Endpoint string
	IPv6Endpoint string

	// The http.Client used by this registry
	Underlying *http.Client

	asSetup  sync.Once
	dnsSetup sync.Once
	ipSetup  sync.Once
	setup    sync.Once
}

// RFC7484 Section 3
// Per [RFC7258], in each array of base RDAP URLs, the secure versions
// of the transport protocol SHOULD be preferred and tried first.  For
// example, if the base RDAP URLs array contains both HTTPS and HTTP
// URLs, the bootstrap client SHOULD try the HTTPS version first.

// RFC7484 Section 3
// Base RDAP URLs MUST have a trailing "/" character because they are
// concatenated to the various segments defined in [RFC7482].

// RFC7484 Section 7
// The registries may not contain the requested value.  In these cases,
// there is no known RDAP server for that requested value, and the
// client SHOULD provide an appropriate error message to the user.

// RFC7484 Section 8
// Clients SHOULD NOT fetch the registry on every RDAP request.  Clients SHOULD
// cache the registry

func (r *Registry) ForDomain(domain string) (string, error) {
	r.dnsSetup.Do(func() {
		if r.DNSEndpoint == "" {
			r.DNSEndpoint = DNSUrl
		}
	})

	// RFC7484 Section 4
	// The domain name's authoritative registration data service is found by
	// doing the label-wise longest match of the target domain name with the
	// domain values in the Entry Arrays in the IANA Bootstrap Service
	// Registry for Domain Name Space.  The match is done per label, from
	// right to left.  If the longest match results in multiple entries,
	// then those entries are considered equivalent.

	// RFC7484 Section 4
	// If a domain RDAP query for a.b.example.com matches both com and
	// example.com entries in the registry, then the longest match applies
	// and the example.com entry is used by the client.

	req, err := http.NewRequest("GET", r.DNSEndpoint, nil)
	if err != nil {
		return "", err // TODO(adam)
	}
	r.fixReqPath(req, r.DNSEndpoint)

	resp, err := r.do(req)
	if err != nil {
		return "", err // TODO(adam)
	}

	// Parse response
	response, err := r.readResponse(resp.Body)
	if err != nil {
		return "", err // TODO(adam)
	}

	// naive lookup
	for i := range response.Services {
		svc := response.Services[i]
		if len(svc) != 2 {
			panic(svc)
		}
		for j := range svc[0] {
			// TOOD(adam): actually match accorind to RFC
			if strings.HasSuffix(domain, svc[0][j]) {
				// TODO(adam): need to sort by HTTPS
				return svc[1][0], nil
			}
		}
	}

	return "", nil
}

func (r *Registry) ForIPNetwork(ip string) (string, error) {
	r.ipSetup.Do(func() {
		if r.IPv4Endpoint == "" {
			r.IPv4Endpoint = IPv4Url
		}
		if r.IPv6Endpoint == "" {
			r.IPv6Endpoint = IPv6Url
		}

	})

	// r.fixReqPath(req, <host>) // TODO

	// RFC7484 Section 5.1
	// For IP address space, the authoritative registration data service is
	// found by doing a longest match of the target address with the values
	// of the arrays in the corresponding RDAP Bootstrap Service Registry
	// for Address Space.  The longest match is done the same way as for
	// routing: the addresses are converted in binary form and then the
	// binary strings are compared to find the longest match up to the
	// specified prefix length.

	// RFC7484 Section 5.2 and 5.3 (same logic)
	// For example, a query for "192.0.2.1/25" matches the "192.0.0.0/8"
	// entry and the "192.0.2.0/24" entry in the example registry above.
	// The latter is chosen by the client given the longest match.

	return "", nil
}

func (r *Registry) ForASNumber(asn string) (string, error) {
	r.asSetup.Do(func() {
		if r.ASNEndpoint == "" {
			r.ASNEndpoint = ASNUrl
		}
	})
	// RFC7484 Section 5.3
	// The array always contains two AS numbers represented in decimal format
	// that represents the range of AS numbers between the two elements of the
	// array. A single AS number is represented as a range of two identical AS
	// numbers.

	// r.fixReqPath(req, r.ASNEndpoint)

	return "", nil
}

func (r *Registry) do(req *http.Request) (*http.Response, error) {
	r.setup.Do(func() {
		if r.Underlying == nil {
			r.Underlying = DefaultHTTPClient
		}
	})

	// Set Accept header
	if v := req.Header.Get("Accept"); v == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Perform the request
	return r.Underlying.Do(req)
}

func (r *Registry) readResponse(rdr io.ReadCloser) (*Response, error) {
	defer rdr.Close()

	bs, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	var resp Response
	if err := json.Unmarshal(bs, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *Registry) fixReqPath(req *http.Request, host string) {
	if req.URL.Host == "" {
		raw := host + req.URL.Path
		u, err := url.Parse(raw)
		if err != nil {
			panic(fmt.Errorf("invalid url %q: %v", raw, err))
		}
		req.URL = u
	}
}
