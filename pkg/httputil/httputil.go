package httputil

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type Config struct {
	InsecureSkipVerify bool
}

func Transport(cfg *Config) http.RoundTripper {
	tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
			TLSHandshakeTimeout:   1 * time.Minute,
			IdleConnTimeout:       1 * time.Minute,
			ResponseHeaderTimeout: 1 * time.Minute,
			ExpectContinueTimeout: 1 * time.Minute,
	}

	if cfg == nil {
		return tr
	}
	// Override known specifics
	if cfg.InsecureSkipVerify {
		tr.TLSClientConfig.InsecureSkipVerify = cfg.InsecureSkipVerify
	}
	return tr
}
