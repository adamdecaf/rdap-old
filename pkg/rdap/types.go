package rdap

import (
	"fmt"
	"strings"
)

// RFC 7483 Section 4.9
// An objectClassName is REQUIRED in all RDAP response objects so that
// the type of the object can be interpreted.

// RFC 7483 Section 5.1
// See rfc-7483-section-5-1-example.json
type Entity struct {
	Handle string
	Title string
	// Role // RFC 7483 Section 10.2.4
	// Addresses []...
	// Telephone []...
	Emails []string
	// Coordinates []..
	// Status

	// "ldhName" : "ns1.example.com" // TODO(adam): Required?
	// a string containing the LDH name of the nameserver (see Section 3)
	//
	// LDH names:        Textual representations of DNS names where the
	// labels of the domain are all "letters, digits,
	// hyphen" labels as described by [RFC5890].  Trailing
	// periods are optional.

	// "objectClassName" : "entity"
}

// RFC 7483 Section 5.2
// See rfc-7483-section-5-2-example.json
type Nameserver struct {
	// objectClassName -- the string "nameserver"

	Handle string
	// Status
	// IPV4Addresses []net.IP
	// IPV6Addresses []...
}

// RFC 7483 Section 5.3
// See rfc-7483-section-5-3-example.json
type Domain struct {
	// objectClassName -- the string "domain"

	//   handle -- a string representing a registry unique identifier of
	// the domain object instance

	//   ldhName -- a string describing a domain name in LDH form as
	// described in Section 3

	// o  secureDNS -- an object with the following members:
	//   *  zoneSigned -- true if the zone has been signed, false
	//      otherwise.

	// entities -- an array of entity objects as defined by Section 5.1

	// status -- see Section 4.6

	// network -- represents the IP network for which a reverse DNS
	// domain is referenced.  See Section 5.4
}

// RFC 7483 Section 5.4
// See rfc-7483-section-5-4-example.json
type IPNetwork struct {
	// "objectClassName" : "ip network",
	// "handle" : "XXXX-RIR",
	// "startAddress" : "2001:db8::",
	// "endAddress" : "2001:db8:0:ffff:ffff:ffff:ffff:ffff",

	// "ipVersion" : "v6",
	// ipVersion -- a string signifying the IP protocol version of the
	// network: "v4" signifies an IPv4 network, and "v6" signifies an
	// IPv6 network

	// "name": "NET-RTR-1",

	// "type" : "DIRECT ALLOCATION",
	// type -- a string containing an RIR-specific classification of the
	// network

	// "country" : "AU",
	// status
}

// RFC 7483 Section 5.5
// See rfc-7483-section-5-5-example.json
type Autnum struct {
	// objectClassName -- the string "autnum"

	// handle -- a string representing an RIR-unique identifier of the
	// autnum registration

	// o  startAutnum -- a number representing the starting number [RFC5396]
	// in the block of Autonomous System numbers

	// o  endAutnum -- a number representing the ending number [RFC5396] in
	// the block of Autonomous System numbers

	// o  name -- an identifier assigned to the autnum registration by the
	// registration holder

	// o  type -- a string containing an RIR-specific classification of the
	// autnum

	// o  status -- an array of strings indicating the state of the autnum

	// o  country -- a string containing the name of the two-character
	// country code of the autnum
}

// RFC 7483 Section 6
// See rfc-7483-section-6-example.json
type Error struct {
	Code int `json:"errorCode"`
	Title string `json:"title"`
	Description []string `json:"description"`
}

func (e *Error) Error() string {
	if len(e.Description) > 0 {
		return fmt.Sprintf("%s (Code: %d): %s", e.Title, e.Code, strings.Join(e.Description, " "))
	}
	return fmt.Sprintf("%s (Code: %d)", e.Title, e.Code)
}
