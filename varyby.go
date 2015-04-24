package webrate

import (
	"net"
	"net/http"
)

// VaryByIP returns a custom "vary by" function for throttled, which varies
// request based on client's IP address from the given header.  If header is
// empty, extracts IP address from r.RemoteAddr.
func VaryByIP(headerName string) func(r *http.Request) string {
	return func(r *http.Request) string {
		return getRequestIP(r, headerName)
	}
}

// VaryByPathAndIP is like VaryByIP but also adds request path.
func VaryByPathAndIP(headerName string) func(r *http.Request) string {
	return func(r *http.Request) string {
		return r.URL.Path + "\n" + getRequestIP(r, headerName)
	}
}

// getRequestIP returns a remote IP address of the client that made the
// request. The address is take from the given header name, or from RemoteAddr
// if the header name is an empty string.
func getRequestIP(r *http.Request, headerName string) string {
	if headerName == "" {
		return extractIP(r.RemoteAddr)
	}
	return extractIP(r.Header.Get(headerName))
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
