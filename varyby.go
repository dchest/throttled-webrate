package webrate

import (
	"net"
	"net/http"
)

// VaryByIP is a custom "vary by" function for throttled, which varies request
// based on client's IP address from the given header.  If header is empty,
// extracts IP address from r.RemoteAddr.
func VaryByIP(r *http.Request, headerName string) string {
	if headerName == "" {
		return extractIP(r.RemoteAddr)
	}
	return extractIP(r.Header.Get(headerName))
}

// VaryByPathAndIP is like VaryByIP but also adds request path.
func VaryByPathAndIP(r *http.Request, headerName string) string {
	return r.URL.Path + "\n" + VaryByIP(r, headerName)
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
