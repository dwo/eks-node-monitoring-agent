package dcgm

import (
	"context"
	"fmt"
	"net"
	"net/netip"
)

type ipResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// resolveDCGMAddress gives DCGM a literal address because its hostname path requests IPv4 only.
// https://github.com/NVIDIA/DCGM/blob/72fa3feaa67d716a75323a8f47c34ff3ee73f824/common/transport/DcgmIpc.cpp#L773-L787
func resolveDCGMAddress(ctx context.Context, resolver ipResolver, address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("parse %q as host:port: %w", address, err)
	}

	if ip, err := netip.ParseAddr(host); err == nil {
		return net.JoinHostPort(ip.String(), port), nil
	}

	addresses, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return "", fmt.Errorf("resolve DCGM host %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return "", fmt.Errorf("resolve DCGM host %q: no IP addresses found", host)
	}

	// Preserve the DCGM client's existing IPv4 preference while supporting
	// hostnames that resolve exclusively to IPv6 addresses.
	selected := addresses[0]
	for _, candidate := range addresses {
		if candidate.Is4() {
			selected = candidate
			break
		}
	}

	return net.JoinHostPort(selected.String(), port), nil
}
