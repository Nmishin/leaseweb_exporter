package collector

import (
	"context"
	"strings"
	"testing"

	"github.com/Nmishin/leaseweb_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeFloatingIPs struct {
	defs []client.FloatingIPDefinition
}

func (f fakeFloatingIPs) ListRanges(context.Context) ([]client.FloatingIPRange, error) {
	return []client.FloatingIPRange{{ID: "81.17.54.216_31"}}, nil
}

func (f fakeFloatingIPs) ListDefinitions(context.Context, string) ([]client.FloatingIPDefinition, error) {
	return f.defs, nil
}

func def(ip, anchor, passive string) client.FloatingIPDefinition {
	return client.FloatingIPDefinition{
		RangeID: "81.17.54.216_31", FloatingIP: ip,
		AnchorIP: anchor, PassiveAnchorIP: passive, Status: "ACTIVE",
		UpdatedAt: "2026-10-05T15:52:24+00:00",
	}
}

func TestFloatingIPRangeAnchorIPs(t *testing.T) {
	tests := []struct {
		name string
		defs []client.FloatingIPDefinition
		want string
	}{
		{
			name: "two IPs on two servers",
			defs: []client.FloatingIPDefinition{
				def("81.17.54.216/32", "95.211.61.194", "95.211.61.193"),
				def("81.17.54.217/32", "95.211.61.193", "95.211.61.194"),
			},
			want: "2",
		},
		{
			name: "both IPs collapsed onto one server",
			defs: []client.FloatingIPDefinition{
				def("81.17.54.216/32", "95.211.61.193", "95.211.61.194"),
				def("81.17.54.217/32", "95.211.61.193", "95.211.61.194"),
			},
			want: "1",
		},
		{
			name: "collapsed with passive anchors following the active one",
			defs: []client.FloatingIPDefinition{
				def("81.17.54.216/32", "95.211.61.193", "95.211.61.193"),
				def("81.17.54.217/32", "95.211.61.193", "95.211.61.193"),
			},
			want: "1",
		},
		{
			name: "single IP",
			defs: []client.FloatingIPDefinition{
				def("81.17.54.216/32", "95.211.61.193", "95.211.61.194"),
			},
			want: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFloatingIPCollector(fakeFloatingIPs{defs: tt.defs})
			expected := `
# HELP leaseweb_floating_ip_range_anchor_ips Number of distinct anchor IPs currently used by floating IPs in the range
# TYPE leaseweb_floating_ip_range_anchor_ips gauge
leaseweb_floating_ip_range_anchor_ips{range_id="81.17.54.216_31"} ` + tt.want + "\n"
			if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "leaseweb_floating_ip_range_anchor_ips"); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestFloatingIPRangeHasNoSpreadMetrics(t *testing.T) {
	c := NewFloatingIPCollector(fakeFloatingIPs{defs: []client.FloatingIPDefinition{
		def("81.17.54.216/32", "95.211.61.193", "95.211.61.194"),
	}})
	for _, name := range []string{
		"leaseweb_floating_ip_range_expected_anchor_ips",
		"leaseweb_floating_ip_range_anchors_spread",
	} {
		if n := testutil.CollectAndCount(c, name); n != 0 {
			t.Errorf("%s: got %d series, want 0", name, n)
		}
	}
}
