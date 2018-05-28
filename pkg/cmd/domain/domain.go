package domain

import (
	"fmt"

	"github.com/adamdecaf/rdap/pkg/cmd"
	"github.com/adamdecaf/rdap/pkg/httputil"
	"github.com/adamdecaf/rdap/pkg/rdap"
	"github.com/adamdecaf/rdap/pkg/rdap/bootstrap"
)

func PrintDetails(cfg *cmd.Config, d string) error {
	boot := bootstrap.Registry{}
	if cfg.InsecureSkipVerify {
		bootstrap.DefaultHTTPClient.Transport = httputil.Transport(&httputil.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		})
	}

	server, err := boot.ForDomain(d)
	if err != nil {
		return fmt.Errorf("getting boot strap files: %v", err)
	}
	if server == "" {
		return fmt.Errorf("no server found for %s", d)
	}

	client := rdap.Client{
		BaseAddress: server,
		Debug: cfg.Debug,
	}
	if cfg.InsecureSkipVerify {
		rdap.DefaultHTTPClient.Transport = httputil.Transport(&httputil.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		})
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
