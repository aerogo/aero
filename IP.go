// Copyright (c) 2018 Shen Sheng
// Copyright (c) 2019 Eduard Urbach

package aero

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// CIDR = Classless Inter-Domain Routing.
// Here we'll get the private CIDR blocks.
var privateCIDRs = getPrivateCIDRs()

// isPrivateAddress checks if the address is under private CIDR blocks.
func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)

	if ipAddress == nil {
		return false, errors.New("Address is not valid")
	}

	for _, cidr := range privateCIDRs {
		if cidr.Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

// getPrivateCIDRs returns the list of private CIDR blocks.
func getPrivateCIDRs() []*net.IPNet {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	cidrs := make([]*net.IPNet, len(maxCidrBlocks))

	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}

	return cidrs
}

// realIP returns the client's real public IP address from http request headers.
func realIP(r *http.Request) string {
	xForwardedFor := r.Header.Get(forwardedForHeader)

	// Check the list of IP addresses in the X-Forwarded-For
	// header and return the first global address, if available.
	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := isPrivateAddress(address)

		if err == nil && !isPrivate {
			return address
		}
	}

	// Return the X-Real-Ip header, if available.
	xRealIP := r.Header.Get(realIPHeader)

	if xRealIP != "" {
		return xRealIP
	}

	// If both headers failed, return the remote IP.
	remoteIP := r.RemoteAddr

	// If there is a colon in the remote address,
	// remove the port number.
	if strings.ContainsRune(remoteIP, ':') {
		remoteIP, _, _ = net.SplitHostPort(remoteIP)
	}

	return remoteIP
}
