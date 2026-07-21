package dcgm

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeIPResolver struct {
	addresses []netip.Addr
	err       error
}

func (r fakeIPResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.addresses, r.err
}

func TestResolveDCGMAddress(t *testing.T) {
	tests := map[string]struct {
		address    string
		resolver   fakeIPResolver
		expected   string
		errMessage string
	}{
		"IPv4 literal": {
			address:  "192.0.2.1:5555",
			expected: "192.0.2.1:5555",
		},
		"IPv6 literal": {
			address:  "[2001:db8::1]:5555",
			expected: "[2001:db8::1]:5555",
		},
		"IPv4 hostname": {
			address:  "dcgm.example:5555",
			resolver: fakeIPResolver{addresses: []netip.Addr{netip.MustParseAddr("192.0.2.2")}},
			expected: "192.0.2.2:5555",
		},
		"IPv6 hostname": {
			address:  "dcgm.example:5555",
			resolver: fakeIPResolver{addresses: []netip.Addr{netip.MustParseAddr("2001:db8::2")}},
			expected: "[2001:db8::2]:5555",
		},
		"dual-stack hostname": {
			address: "dcgm.example:5555",
			resolver: fakeIPResolver{addresses: []netip.Addr{
				netip.MustParseAddr("2001:db8::3"),
				netip.MustParseAddr("192.0.2.3"),
			}},
			expected: "192.0.2.3:5555",
		},
		"invalid address": {
			address:    "dcgm.example",
			errMessage: "parse \"dcgm.example\" as host:port",
		},
		"lookup failure": {
			address:    "dcgm.example:5555",
			resolver:   fakeIPResolver{err: errors.New("lookup failed")},
			errMessage: "resolve DCGM host \"dcgm.example\": lookup failed",
		},
		"no addresses": {
			address:    "dcgm.example:5555",
			errMessage: "resolve DCGM host \"dcgm.example\": no IP addresses found",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := resolveDCGMAddress(context.Background(), test.resolver, test.address)
			if test.errMessage != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errMessage)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
