package coredns

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/qaware/minikube-support/pb"
)

// server is a small grpc service that answers to dns queries over grpc from CoreDNS.
// Please refer to the CoreDNS GRPC Plugin how to configure it to use this as backend.
type server struct {
	entries     map[dns.Type]map[dns.Name][]dns.RR
	entriesLock sync.RWMutex
	server      *grpc.Server
}

// NewServer initializes the grpc core dns service.
func NewServer() *server {
	return &server{
		entries:     make(map[dns.Type]map[dns.Name][]dns.RR),
		entriesLock: sync.RWMutex{},
	}
}

// Start starts the server by using the given socket to listen on for new dns queries.
func (srv *server) Start(socket net.Listener) {
	srv.server = grpc.NewServer()
	pb.RegisterDnsServiceServer(srv.server, srv)

	go func() {
		e := srv.server.Serve(socket)
		if e != nil {
			logrus.Errorf("unable to start serving dns requests: %s", e)
		}
	}()
}

// Stop tries to stop the server gracefully.
// It enforces the server to stop if the server can not be stopped gracefully within 5 seconds.
func (srv *server) Stop() {
	defer srv.server.Stop()
	go srv.server.GracefulStop()
	time.Sleep(5 * time.Second)
}

// Query answers to dns queries received via grpc from CoreDNS.
// See also https://coredns.io/plugins/grpc/ for information about the configuration
// of CoreDNS to use this as backend.
func (srv *server) Query(ctx context.Context, in *pb.DnsPacket) (*pb.DnsPacket, error) {
	m := new(dns.Msg)
	if err := m.Unpack(in.Msg); err != nil {
		return nil, fmt.Errorf("failed to unpack msg: %v", err)
	}
	r := new(dns.Msg)
	r.SetReply(m)
	r.Authoritative = true
	for _, q := range r.Question {
		rr, e := srv.GetResourceRecord(dns.Name(q.Name), dns.Type(q.Qtype))
		logrus.Infof("Request for record %s %s", q.Name, dns.Type(q.Qtype))
		if e != nil {
			logrus.Debugf("Can not handle request %v: %s", q, e)
			continue
		}
		logrus.Infof("Found DNS record for %s %s: %s", q.Name, dns.Type(q.Qtype), rr)
		r.Answer = append(r.Answer, rr...)
	}

	if len(r.Answer) == 0 {
		r.Rcode = dns.RcodeNameError
	}

	out, err := r.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack msg: %v", err)
	}
	return &pb.DnsPacket{Msg: out}, nil
}

// AddHost adds the given domain name as new resource record. Depending on the given
// ipAddress either as A record or as AAAA record.
func (srv *server) AddHost(name string, ipAddress string) error {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("can not parse ip: %s", ipAddress)
	}

	if IsIPv4(ip) {
		return srv.AddA(name, ip)
	} else {
		return srv.AddAAAA(name, ip)
	}
}

// AddA adds a new A resource record for the given domain to the internal database.
// It will not overwrite any existing resource records if there is already one with the same name and type.
func (srv *server) AddA(name string, ipv4 net.IP) error {
	if ipv4 == nil {
		return fmt.Errorf("given ip address is nil")
	}

	if err := validateDomainName(name); err != nil {
		return err
	}

	if !IsIPv4(ipv4) {
		return fmt.Errorf("given IP %s is not an IPv4 address", ipv4)
	}

	srv.addRR(&dns.A{
		Hdr: dns.RR_Header{
			Name:   normalizeName(name),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    10,
		},
		A: ipv4,
	})
	return nil
}

// AddAAAA adds a new AAAA resource record for the given domain to the internal database.
// It will not overwrite any existing resource records if there is already one with the same name and type.
func (srv *server) AddAAAA(name string, ipv6 net.IP) error {
	if ipv6 == nil {
		return fmt.Errorf("given ip address is nil")
	}

	if err := validateDomainName(name); err != nil {
		return err
	}

	if !IsIPv6(ipv6) {
		return fmt.Errorf("given IP %s is not an IPv6 address", ipv6)
	}

	srv.addRR(&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   normalizeName(name),
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    10,
		},
		AAAA: ipv6,
	})
	return nil
}

// AddCNAME adds a new CNAME resource record for the given domain to the internal database.
// It will not overwrite any existing resource records if there is already one with the same name and type.
func (srv *server) AddCNAME(name string, target string) error {
	if err := validateDomainName(name); err != nil {
		return err
	}

	if !dns.IsFqdn(target) {
		logrus.Warnf("%s is not a full qualified domain name", target)
		target = normalizeName(target)
	}

	srv.addRR(&dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   normalizeName(name),
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    10,
		},
		Target: target,
	})
	return nil
}

// addRR adds the given resource record to the internal database.
// It will not overwrite any existing resource records if there is already one with the same name and type.
func (srv *server) addRR(entry dns.RR) {
	srv.entriesLock.Lock()
	defer srv.entriesLock.Unlock()

	name := dns.Name(entry.Header().Name)
	dnsType := dns.Type(entry.Header().Rrtype)
	if _, ok := srv.entries[dnsType]; !ok {
		srv.entries[dnsType] = make(map[dns.Name][]dns.RR)
	}
	srv.entries[dnsType][name] = append(srv.entries[dnsType][name], entry)
	logrus.Infof("Resource Record %s added", entry)
}

// GetResourceRecord tries to find a resource record with the given name and type.
// It will return an error if no records are found.
func (srv *server) GetResourceRecord(name dns.Name, dnsType dns.Type) ([]dns.RR, error) {
	srv.entriesLock.RLock()
	defer srv.entriesLock.RUnlock()

	typeRRs, ok := srv.entries[dnsType]
	if !ok {
		return nil, fmt.Errorf("no resource records of type %s", dnsType)
	}

	rr, ok := typeRRs[name]
	if !ok {
		return nil, fmt.Errorf("no resource record %s with name %s", dnsType, name)
	}
	return rr, nil
}

// RemoveResourceRecord deletes the resource record identified by the name and type from the internal database.
func (srv *server) RemoveResourceRecord(name string, dnsType dns.Type) {
	srv.entriesLock.Lock()
	defer srv.entriesLock.Unlock()

	normalizedName := dns.Name(normalizeName(name))
	records := srv.entries[dnsType]
	delete(records, normalizedName)
	if len(records) == 0 {
		delete(srv.entries, dnsType)
	}
}

// ListRRs returns a list of all currently stored resource records.
// The returned list can be empty if no records are stored.
func (srv *server) ListRRs() []dns.RR {
	srv.entriesLock.RLock()
	defer srv.entriesLock.RUnlock()

	var rrs []dns.RR

	for _, typeRRs := range srv.entries {
		for _, rr := range typeRRs {
			rrs = append(rrs, rr...)
		}
	}
	return rrs
}

// IsIPv4 checks if the given ip address is an IPv4 address.
// It returns true if it is a IPv4 address. Otherwise false.
func IsIPv4(ip net.IP) bool {
	return ip != nil && ip.To4() != nil
}

// IsIPv6 checks if the given ip address is an IPv6 address.
// It returns true if it is a IPv6 address. Otherwise false.
func IsIPv6(ip net.IP) bool {
	return !IsIPv4(ip)
}

func normalizeName(name string) string {
	if dns.IsFqdn(name) {
		return name
	} else {
		return name + "."
	}
}

// validateDomainName check if the given name is valid domain name. If it is valid it returns nil. Otherwise an error.
func validateDomainName(domain string) error {
	if _, ok := dns.IsDomainName(domain); !ok {
		return fmt.Errorf("%s is not a valid domain name", domain)
	}
	return nil
}
