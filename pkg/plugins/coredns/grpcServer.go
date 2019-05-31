package coredns

import (
	"context"
	"fmt"
	"github.com/chr-fritz/minikube-support/pb"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"time"
)

// Server is a small grpc service that answers to dns queries over grpc from CoreDNS.
// Please refer to the CoreDNS GRPC Plugin how to configure it to use this as backend.
type Server struct {
	entries map[dns.Type]map[dns.Name][]dns.RR
	server  *grpc.Server
}

// NewServer initializes the grpc core dns service.
func NewServer() *Server {
	return &Server{
		entries: make(map[dns.Type]map[dns.Name][]dns.RR),
	}
}

// Start starts the server by using the given socket to listen on for new dns queries.
func (srv *Server) Start(socket net.Listener) {
	srv.server = grpc.NewServer()
	pb.RegisterDnsServiceServer(srv.server, &Server{})

	go func() {
		e := srv.server.Serve(socket)
		if e != nil {
			logrus.Errorf("unable to start serving dns requests: %s", e)
		}
	}()
}

// Stop tries to stop the server gracefully.
// It enforces the server to stop if the server can not be stopped gracefully within 5 seconds.
func (srv *Server) Stop() {
	defer srv.server.Stop()
	go srv.server.GracefulStop()
	time.Sleep(5 * time.Second)
}

// Query answers to dns queries received via grpc from CoreDNS.
// See also https://coredns.io/plugins/grpc/ for information about the configuration
// of CoreDNS to use this as backend.
func (srv *Server) Query(ctx context.Context, in *pb.DnsPacket) (*pb.DnsPacket, error) {
	m := new(dns.Msg)
	if err := m.Unpack(in.Msg); err != nil {
		return nil, fmt.Errorf("failed to unpack msg: %v", err)
	}
	r := new(dns.Msg)
	r.SetReply(m)
	r.Authoritative = true

	for _, q := range r.Question {
		rr, e := srv.GetResourceRecord(dns.Name(q.Name), dns.Type(q.Qtype))
		if e != nil {
		}
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
func (srv *Server) AddHost(name string, ipAddress string) error {
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
func (srv *Server) AddA(name string, ipv4 net.IP) error {
	if ipv4 == nil {
		return fmt.Errorf("given ip address is nil")
	}

	if _, ok := dns.IsDomainName(name); !ok {
		return fmt.Errorf("%s is not a valid domain name", name)
	}

	if !IsIPv4(ipv4) {
		return fmt.Errorf("given IP %s is not an IPv4 address", ipv4)
	}

	srv.addRR(&dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
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
func (srv *Server) AddAAAA(name string, ipv6 net.IP) error {
	if ipv6 == nil {
		return fmt.Errorf("given ip address is nil")
	}

	if _, ok := dns.IsDomainName(name); !ok {
		return fmt.Errorf("%s is not a valid domain name", name)
	}

	if !IsIPv6(ipv6) {
		return fmt.Errorf("given IP %s is not an IPv6 address", ipv6)
	}

	srv.addRR(&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   name,
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
func (srv *Server) AddCNAME(name string, target string) error {
	if _, ok := dns.IsDomainName(name); !ok {
		return fmt.Errorf("%s is not a valid domain name", name)
	}

	if !dns.IsFqdn(target) {
		return fmt.Errorf("%s is not a full qualified domain name", target)
	}

	srv.addRR(&dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   name,
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
func (srv *Server) addRR(entry dns.RR) {
	name := dns.Name(entry.Header().Name)
	dnsType := dns.Type(entry.Header().Rrtype)
	if _, ok := srv.entries[dnsType]; !ok {
		srv.entries[dnsType] = make(map[dns.Name][]dns.RR)
	}
	srv.entries[dnsType][name] = append(srv.entries[dnsType][name], entry)
}

// GetResourceRecord tries to find a resource record with the given name and type.
// It will return an error if no records are found.
func (srv *Server) GetResourceRecord(name dns.Name, dnsType dns.Type) ([]dns.RR, error) {
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
func (srv *Server) RemoveResourceRecord(name dns.Name, dnsType dns.Type) {
	delete(srv.entries[dnsType], name)
	if len(srv.entries[dnsType]) == 0 {
		delete(srv.entries, dnsType)
	}
}

// ListRRs returns a list of all currently stored resource records.
// The returned list can be empty if no records are stored.
func (srv *Server) ListRRs() []dns.RR {
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
