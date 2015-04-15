package webrate

import (
	"net"
	"net/http"
)

// VaryByIP is a custom "vary by" function for throttled, which varies request
// based on client's IP address and X-Real-IP header.
func VaryByIP(r *http.Request) string {
	return extractIP(r.RemoteAddr) + "\n" + r.Header.Get("X-Real-IP")
}

// VaryByPathAndIP is like VaryByIP but also adds request path.
func VaryByPathAndIP(r *http.Request) string {
	return r.URL.Path + "\n" + VaryByIP(r)
}

// extractIP extracts IP address (or host) from the given string,
// which may have host and port in it.
func extractIP(addr string) string {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return ip
}
