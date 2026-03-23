package httputil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// LookupIPFunc is a variable to allow mocking net.LookupIP in tests across packages.
var LookupIPFunc = net.LookupIP

// AllowLocalIPsForTesting can be set to true in tests to bypass SSRF IP checks
var AllowLocalIPsForTesting = false

func IsBlockedIP(ip net.IP) bool {
	if AllowLocalIPsForTesting {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// ValidateURL checks if a given URL string is safe from SSRF attacks.
// It explicitly blocks loopback, private, unspecified, and link-local IP addresses.
// It fails closed on DNS resolution errors.
func ValidateURL(u string) error {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("invalid URL scheme")
	}

	host := parsedURL.Hostname()
	if host == "" {
		return errors.New("URL must contain a host")
	}

	ips, err := LookupIPFunc(host)
	if err != nil {
		// Fail closed on DNS resolution error
		return errors.New("DNS resolution failed")
	}

	for _, ip := range ips {
		if IsBlockedIP(ip) {
			return errors.New("URL resolves to a blocked IP address")
		}
	}

	return nil
}

// InitSafeHTTPClient returns an http.Client with a custom DialContext that prevents
// DNS rebinding (TOCTOU) attacks by pinning the connection to the validated IP.
func InitSafeHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			ips, err := LookupIPFunc(host)
			if err != nil {
				return nil, fmt.Errorf("DNS resolution failed: %w", err)
			}
			if len(ips) == 0 {
				return nil, errors.New("no IP addresses found for host")
			}

			// Validate all resolved IPs
			for _, ip := range ips {
				if IsBlockedIP(ip) {
					return nil, errors.New("URL resolves to a blocked IP address")
				}
			}

			// Connect directly to the first validated IP
			safeAddr := net.JoinHostPort(ips[0].String(), port)
			return dialer.DialContext(ctx, network, safeAddr)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
}

// SafeClient is a globally shared http.Client that enforces TOCTOU prevention.
var SafeClient = InitSafeHTTPClient()
