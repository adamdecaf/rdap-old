package main

import (
	"fmt"
	"errors"
	"os"
	"strings"

	"github.com/adamdecaf/rdap/pkg/cmd/domain"
)

const Version = "0.1.0-dev"

// TODO(adam): -f (format) flag for output, similar to docker / go templates

type command struct {
	// args is the os.Args after subcommand
	fn func(args []string) error
	help string
}

func main() {
	commands := make(map[string]*command, 0)
	commands["domain"] = &command{
		fn: func(args []string) error {
			if len(args) == 0 {
				return errors.New("no domain specified")
			}
			return domain.PrintDetails(args[0])
		},
		help: "wat",
	}
	commands["version"] = &command{
		fn: func(_ []string) error {
			fmt.Println(Version)
			return nil
		},
	}

	if len(os.Args) == 1 {
		fmt.Println("ERROR: You must specify a command")
		os.Exit(1)
	}
	raw := strings.ToLower(os.Args[1])
	if cmd, ok := commands[raw]; ok {
		rest := os.Args[2:]
		if err := cmd.fn(rest); err != nil {
			fmt.Printf("ERROR: %s: %v\n", raw, err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("command %s not found\n", raw)
		os.Exit(1)
	}
	os.Exit(0)

	// Allow any http.Client instance
	// RFC 7481 Section 3.2
	// "Clients MUST support both (Basic or Digest auth) to interoperate with
	// servers that support one or the other."
	// Clients can auth with X.509 certificates, SAML, OpenID, OAuth, etc..

	// RFC 7481 Section 3.5
	// "It is also possible to encrypt discrete objects (such as command path
	// segments and JSON-encoded response objects) at one endpoint"
	// offer body as io.Reader ?

	// RFC 7483 Section 5.1
	// The entity object class uses jCard [RFC7095] to represent contact
	// information, such as postal addresses, email addresses, phone numbers
	// and names of organizations and individuals.
}

// RFC 7482 Section
// "Servers MUST return an HTTP 501 (Not Implemented) [RFC7231] response to
// inform clients ofunsupported query types."

// JSON

// RFC 7483 Section 2.1
// Clients processing JSON responses need to be prepared for members
// representing registration data specified in this document to be
// absent from a response.  In other words, servers are free to not
// include JSON members containing registration data based on their own
// policies.
//
// Insertion of unrecognized members ignored by clients may also be used
// for future revisions to this specification.
//
// Finally, all JSON names specified in this document are case
// sensitive.  Both servers and clients MUST transmit and process them
// using the specified character case.

// 4.2.  Links
//
//    The "links" array is found in data structures to signify links to
//    other resources on the Internet.  The relationship of these links is
//    defined by the IANA registry described by [RFC5988].
//
//    The following is an example of the link structure:
//
//        {
//          "value" : "http://example.com/context_uri",
//          "rel" : "self",
//          "href" : "http://example.com/target_uri",
//          "hreflang" : [ "en", "ch" ],
//          "title" : "title",
//          "media" : "screen",
//          "type" : "application/json"
//        }

// 4.6.  Status
//    This data structure, named "status", is an array of strings
//    indicating the state of a registered object (see Section 10.2.2 for a
//    list of values).
//
// Value: validated
// Type: status
// Description: Signifies that the data of the object instance has
//    been found to be accurate.  This type of status is usually
//    found on entity object instances to note the validity of
//    identifying contact information.
//
// Value: renew prohibited
// Type: status
// Description: Renewal or reregistration of the object instance is
//    forbidden.
//
// Value: update prohibited
// Type: status
// Description: Updates to the object instance are forbidden.
//
// Value: transfer prohibited
// Type: status
// Description: Transfers of the registration from one registrar to
//    another are forbidden.  This type of status normally applies to
//    DNR domain names.
//
// Value: delete prohibited
// Type: status
// Description: Deletion of the registration of the object instance
//    is forbidden.  This type of status normally applies to DNR
//    domain names.
//
// Value: proxy
// Type: status
// Description: The registration of the object instance has been
//    performed by a third party.  This is most commonly applied to
//    entities.
//
// Value: private
// Type: status
// Description: The information of the object instance is not
//    designated for public consumption.  This is most commonly
//    applied to entities.
//
// Value: removed
// Type: status
// Description: Some of the information of the object instance has
//    not been made available and has been removed.  This is most
//    commonly applied to entities.
//
// Value: obscured
// Type: status
// Description: Some of the information of the object instance has
//    been altered for the purposes of not readily revealing the
//    actual information of the object instance.  This is most
//    commonly applied to entities.
//
// Value: associated
// Type: status
// Description: The object instance is associated with other object
//    instances in the registry.  This is most commonly used to
//    signify that a nameserver is associated with a domain or that
//    an entity is associated with a network resource or domain.
//
// Value: active
// Type: status
// Description: The object instance is in use.  For domain names, it
//    signifies that the domain name is published in DNS.  For
//    network and autnum registrations, it signifies that they are
//    allocated or assigned for use in operational networks.  This
//    maps to the "OK" status of the Extensible Provisioning Protocol
//    (EPP) [RFC5730] .
//
// Value: inactive
// Type: status
// Description: The object instance is not in use.  See "active".
//
//
// Value: locked
// Type: status
// Description: Changes to the object instance cannot be made,
//    including the association of other object instances.
//
// Value: pending create
// Type: status
// Description: A request has been received for the creation of the
//    object instance, but this action is not yet complete.
//
//
// Value: pending renew
// Type: status
// Description: A request has been received for the renewal of the
//    object instance, but this action is not yet complete.
//
// Value: pending transfer
// Type: status
// Description: A request has been received for the transfer of the
//    object instance, but this action is not yet complete.
//
// Value: pending update
// Type: status
// Description: A request has been received for the update or
//    modification of the object instance, but this action is not yet
//    complete.
//
// Value: pending delete
// Type: status
// Description: A request has been received for the deletion or
//    removal of the object instance, but this action is not yet
//    complete.  For domains, this might mean that the name is no
//    longer published in DNS but has not yet been purged from the
//    registry database.
