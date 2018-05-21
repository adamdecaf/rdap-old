package domain

import (
	"fmt"

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
		// BaseAddress: rdap.DefaultServer,
	}

	resp, err := client.Domain(d)
	if err != nil {
		return fmt.Errorf("grabbing %s: %v", d, err)
	}
	if resp != nil {
		fmt.Println(resp)
	}
	return nil
}