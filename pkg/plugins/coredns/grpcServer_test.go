package coredns

import (
	"net"
	"reflect"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		want *server
	}{
		{"create server", &server{entries: make(map[dns.Type]map[dns.Name][]dns.RR)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_AddHost(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		ipAddress   string
		wantErr     bool
		recordFound bool
		wantType    uint16
	}{
		{"ok ipv4", "ok.", "192.168.1.1", false, true, dns.TypeA},
		{"ok ipv6", "ok.", "::1", false, true, dns.TypeAAAA},
		{"to-long domain", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", "192.168.1.1", true, false, dns.TypeA},
		{"invalid ip 1", "invalid-ip", "192.168.1.1.1", true, false, dns.TypeA},
		{"invalid ip 2", "invalid-ip", "127.0..0.1", true, false, dns.TypeA},
		{"invalid ip 3", "invalid-ip", "abcdefgh.jkl.1.1", true, false, dns.TypeA},
		{"invalid ip 4", "invalid-ip", "abcdef:jkl::", true, false, dns.TypeA},
		{"invalid ip 5", "invalid-ip", "fd00::1::1", true, false, dns.TypeAAAA},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			if err := srv.AddHost(tt.domain, tt.ipAddress); (err != nil) != tt.wantErr {
				t.Errorf("server.AddHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rr := checkResourceRecord(t, srv, tt.domain, dns.Type(tt.wantType), tt.recordFound)
			if rr == nil {
				return
			}
			assert.Equal(t, tt.domain, rr.Header().Name)
			assert.Equal(t, tt.wantType, rr.Header().Rrtype)
		})
	}
}

func TestServer_AddA(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		ipv4        string
		wantErr     bool
		recordFound bool
	}{
		{"ok", "ok.", "192.168.1.1", false, true},
		{"to-long domain", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", "192.168.1.1", true, false},
		{"invalid ip 1", "invalid-ip", "192.168.1.1.1", true, false},
		{"invalid ip 3", "invalid-ip", "abcdefgh.jkl.1.1", true, false},
		{"ip is ipv6", "invalid-ip", "ff00::1", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			if err := srv.AddA(tt.domain, net.ParseIP(tt.ipv4)); (err != nil) != tt.wantErr {
				t.Errorf("server.AddA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			rr := checkResourceRecord(t, srv, tt.domain, dns.Type(dns.TypeA), tt.recordFound)
			if rr == nil {
				return
			}
			a, ok := rr.(*dns.A)
			if !ok {
				t.Errorf("Found resource record is not a A record: got %s", reflect.TypeOf(rr))
				return
			}
			assert.Equal(t, tt.domain, a.Header().Name)
			assert.Equal(t, tt.ipv4, a.A.String())
		})
	}
}

func TestServer_AddAAAA(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		ipv6        string
		wantErr     bool
		recordFound bool
	}{
		{"ok", "ok.", "::1", false, true},
		{"to-long domain", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", "::1", true, false},
		{"invalid ip-1", "invalid-ip", "127.0..0.1", true, false},
		{"invalid ip-2", "invalid-ip", "abcdef:jkl::", true, false},
		{"invalid ip-3", "invalid-ip", "fd00::1::1", true, false},
		{"ip is ipv4", "invalid-ip", "127.0.0.1", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			if err := srv.AddAAAA(tt.domain, net.ParseIP(tt.ipv6)); (err != nil) != tt.wantErr {
				t.Errorf("server.AddAAAA() error = %v, wantErr %v", err, tt.wantErr)
			}

			rr := checkResourceRecord(t, srv, tt.domain, dns.Type(dns.TypeAAAA), tt.recordFound)
			if rr == nil {
				return
			}
			a, ok := rr.(*dns.AAAA)
			if !ok {
				t.Errorf("Found resource record is not a A record: got %s", reflect.TypeOf(rr))
				return
			}
			assert.Equal(t, tt.ipv6, a.AAAA.String())
		})
	}
}

func TestServer_AddCNAME(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		target      string
		wantErr     bool
		recordFound bool
		wantTarget  string
	}{
		{"ok", "ok.", "target.example.com.", false, true, "target.example.com."},
		{"to-long domain", "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", "", true, false, ""},
		{"no fqdn", "no-fqdn.", "no-fqdn", false, true, "no-fqdn."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			if err := srv.AddCNAME(tt.domain, tt.target); (err != nil) != tt.wantErr {
				t.Errorf("server.AddCNAME() error = %v, wantErr %v", err, tt.wantErr)
			}

			rr := checkResourceRecord(t, srv, tt.domain, dns.Type(dns.TypeCNAME), tt.recordFound)
			if rr == nil {
				return
			}
			a, ok := rr.(*dns.CNAME)
			if !ok {
				t.Errorf("Found resource record is not a CNAME record: got %s", reflect.TypeOf(rr))
				return
			}
			assert.Equal(t, tt.wantTarget, a.Target)
		})
	}
}

func TestServer_RemoveResourceRecord(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		dnsType dns.Type
		double  bool
		hasType bool
	}{
		{"remove A but other exists not normalized", "domain", dns.Type(dns.TypeA), false, true},
		{"remove A but other exists", "domain.", dns.Type(dns.TypeA), false, true},
		{"remove A but other exists", "domain.", dns.Type(dns.TypeA), true, true},
		{"remove last AAAA not normalized", "domain", dns.Type(dns.TypeAAAA), false, false},
		{"remove last AAAA", "domain.", dns.Type(dns.TypeAAAA), false, false},
		{"remove last AAAA", "domain.", dns.Type(dns.TypeAAAA), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			assert.NoError(t, srv.AddA("domain.", net.ParseIP("127.0.0.1")))
			assert.NoError(t, srv.AddA("domain1.", net.ParseIP("127.0.0.2")))
			assert.NoError(t, srv.AddAAAA("domain.", net.ParseIP("::1")))
			if tt.double {
				assert.NoError(t, srv.AddA("domain.", net.ParseIP("127.0.0.2")))
				assert.NoError(t, srv.AddAAAA("domain.", net.ParseIP("::2")))
			}

			srv.RemoveResourceRecord(tt.domain, tt.dnsType)
			rrs, ok := srv.entries[tt.dnsType]

			if ok != tt.hasType {
				t.Errorf("Wrong removal of type array: want %v, done %v", tt.hasType, ok)
				return
			}
			if _, ok := rrs[dns.Name(tt.domain)]; ok {
				t.Errorf("Want to remove %s but was found: %v", tt.domain, ok)
			}
		})
	}
}

func TestServer_ListRRs(t *testing.T) {
	srv := NewServer()
	assert.NoError(t, srv.AddA("domain", net.ParseIP("127.0.0.1")))
	assert.NoError(t, srv.AddA("domain", net.ParseIP("127.0.0.2")))
	assert.NoError(t, srv.AddA("domain1", net.ParseIP("127.0.0.2")))
	assert.NoError(t, srv.AddAAAA("domain", net.ParseIP("::1")))
	got := srv.ListRRs()
	assert.Equal(t, 4, len(got))
}

func checkResourceRecord(t *testing.T, srv *server, domain string, rrType dns.Type, recordFound bool) dns.RR {
	rr, e := srv.GetResourceRecord(dns.Name(domain), rrType)
	if (e != nil) == recordFound {
		t.Errorf("No %s resource record found for %s: %s", rrType, domain, e)
		return nil
	}
	if rr == nil {
		return nil
	}

	if len(rr) != 1 {
		t.Errorf("Expect only one resource record. Got %v", len(rr))
		return nil
	}

	assert.Equal(t, domain, rr[0].Header().Name)
	assert.Equal(t, dns.Type(rr[0].Header().Rrtype), rrType)
	return rr[0]
}

func TestServer_GetResourceRecord(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		dnsType dns.Type
		wantErr bool
	}{
		{"A domain", "domain.", dns.Type(dns.TypeA), false},
		{"not found A domain", "not-found", dns.Type(dns.TypeA), true},
		{"AAAA domain", "domain.", dns.Type(dns.TypeAAAA), false},
		{"not found AAAA domain", "not found", dns.Type(dns.TypeAAAA), true},
		{"not found CNAME type", "domain", dns.Type(dns.TypeCNAME), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer()
			assert.NoError(t, srv.AddA("domain.", net.ParseIP("127.0.0.1")))
			assert.NoError(t, srv.AddAAAA("domain.", net.ParseIP("::1")))
			got, err := srv.GetResourceRecord(dns.Name(tt.domain), tt.dnsType)
			if (err != nil) != tt.wantErr {
				t.Errorf("server.GetResourceRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}
			assert.Equal(t, 1, len(got))
			assert.Equal(t, tt.domain, got[0].Header().Name)
			assert.Equal(t, tt.dnsType, dns.Type(got[0].Header().Rrtype))
		})
	}
}
