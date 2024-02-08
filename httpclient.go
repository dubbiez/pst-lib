// httpclient.go

package pstlib

import (
	"log"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Constants related to the HTTP client.
const (
	tcpKeepAliveInterval = 30 * time.Second
	maxRedirectsCount    = 10
	userAgentHeader      = "User-Agent"
	defaultUserAgent     = "Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0)"
)

// Config holds the configuration parameters for the HTTP client.
type Config struct {
	ProxyAddr    string
	ProxyURL    string
	UserAgent    string
	AuthType     int
	Username     string
	Password     string
	IgnoreCert   bool
	Timeout      int
	MaxRedirects int
}

// MakeHTTPClient initializes and returns a new HTTP client configured with a proxy and custom dialer.
func MakeHTTPClient(cfg *Config) *http.Client {
	proxyURL, err := url.Parse(cfg.ProxyAddr)
	if err != nil {
		log.Fatalf("Error parsing proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			KeepAlive: tcpKeepAliveInterval,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.IgnoreCert},
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	if cfg.MaxRedirects <= 0 {
		transport.DisableKeepAlives = true
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return http.ErrUseLastResponse
			}
			// Copy the headers from the initial request to the next request.
			if len(via) > 0 {
				for attr, val := range via[0].Header {
					req.Header[attr] = val
				}
			}
			return nil
		},
	}

	// Add a default User-Agent if one is not provided.
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	return client
}