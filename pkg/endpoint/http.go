package endpoint

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewClient creates http/s client.
// if tlsConfig is nil, it creates http client otherwise, it creates https client
func NewClient(tlsConfig *tls.Config) *http.Client {
	transport := &http.Transport{}
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}
}
