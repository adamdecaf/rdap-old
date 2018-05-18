package domain

import (
	"fmt"
	"net"

	"github.com/adamdecaf/rdap/pkg/rdap"
	"github.com/adamdecaf/rdap/pkg/rdap/bootstrap"
)

func PrintDetails(d string) error {
	boot := bootstrap.Registry{}
	server, err := boot.ForDomain(d)
	if err != nil {
		return fmt.Errorf("getting boot strap files: %v", err)
	}
	if server == "" {
		return fmt.Errorf("no server found for %s", d)
	}

	client := rdap.Client{
		BaseAddress: server,
	}

	addrs, err := net.LookupHost(d)
	if err != nil {
		return fmt.Errorf("resolving %d: %v", d, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("no records found for %s", d)
	}

	resp, err := client.IP(addrs[0])
	if err != nil {
		return fmt.Errorf("grabbing %s: %v", d, err)
	}
	if resp != nil {
		fmt.Println(resp)
	}
	return nil
}
