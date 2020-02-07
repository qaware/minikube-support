package coredns

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/miekg/dns"
	"github.com/qaware/minikube-support/pb"
	"github.com/qaware/minikube-support/pkg/apis"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGrpcPlugin(t *testing.T) {
	type host struct {
		name   string
		target string
		ip     string
	}
	type eResp struct {
		name    string
		target  string
		dnsType uint16
	}
	tests := []struct {
		name      string
		addHost   []host
		query     string
		queryType uint16
		expNumRR  int
		responses []eResp
	}{
		{"single", []host{{"dummy", "", "127.0.0.1"}}, "dummy.", dns.TypeA, 1, []eResp{{"dummy.", "127.0.0.1", dns.TypeA}}},
		{"multiple", []host{{"dummy", "", "127.0.0.1"}, {"dummy", "", "127.0.0.2"}}, "dummy.", dns.TypeA, 2, []eResp{{"dummy.", "127.0.0.1", dns.TypeA}, {"dummy.", "127.0.0.2", dns.TypeA}}},
		{"none", []host{}, "dummy.", dns.TypeA, 0, []eResp{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			plugin, server, conn, e := initClientServer(t)
			if e != nil {
				t.Errorf("Fail to init client and server: %s", e)
				return
			}

			defer func() {
				assert.NoError(t, plugin.Stop())
				assert.NoError(t, conn.Close())
			}()

			c := pb.NewDnsServiceClient(conn)

			for _, h := range tt.addHost {
				if h.ip != "" {
					assert.NoError(t, server.AddHost(h.name, h.ip))
				}
				if h.target != "" {
					assert.NoError(t, server.AddCNAME(h.name, h.target))
				}
			}

			response := query(t, c, tt.query, tt.queryType)
			assert.Equal(t, tt.expNumRR, len(response.Answer))

			for i, rr := range response.Answer {
				r := tt.responses[i]
				switch r.dnsType {
				case dns.TypeA:
					checkA(t, rr, r.name, r.target)
				case dns.TypeAAAA:
					checkAAAA(t, rr, r.name, r.target)
				case dns.TypeCNAME:
					checkCNAME(t, rr, r.name, r.target)
				}
			}
		})
	}
}

func initClientServer(t *testing.T) (apis.StartStopPlugin, *server, *grpc.ClientConn, error) {
	plugin := NewGrpcPlugin()
	channel := make(chan *apis.MonitoringMessage, 100)
	_, e := plugin.Start(channel)

	if e != nil {
		return nil, nil, nil, fmt.Errorf("can not start plugin: %s", e)
	}

	server, e := GetServer(plugin)
	if e != nil {
		return nil, nil, nil, fmt.Errorf("can not find grpc server: %s", e)
	}

	// run client
	conn, err := grpc.Dial("localhost:8053", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}

	return plugin, server, conn, nil
}

func query(t *testing.T, c pb.DnsServiceClient, name string, dnsType uint16) *dns.Msg {
	req := new(dns.Msg)
	req.Question = append(req.Question, dns.Question{Name: name, Qtype: dnsType})
	pack, e := req.Pack()
	assert.NoError(t, e)

	resp, e := c.Query(context.Background(), &pb.DnsPacket{Msg: pack})
	assert.NoError(t, e)

	response := new(dns.Msg)
	assert.NoError(t, response.Unpack(resp.Msg))
	return response
}

func checkA(t *testing.T, rr dns.RR, name string, ip string) {
	assert.Equal(t, name, rr.Header().Name)
	assert.Equal(t, dns.TypeA, rr.Header().Rrtype)

	a, ok := rr.(*dns.A)
	if !ok {
		t.Errorf("Found resource record is not a A record: got %s", reflect.TypeOf(rr))
		return
	}
	assert.Equal(t, ip, a.A.String())
}
func checkAAAA(t *testing.T, rr dns.RR, name string, ip string) {
	assert.Equal(t, name, rr.Header().Name)
	assert.Equal(t, dns.TypeAAAA, rr.Header().Rrtype)

	a, ok := rr.(*dns.AAAA)
	if !ok {
		t.Errorf("Found resource record is not a AAAA record: got %s", reflect.TypeOf(rr))
		return
	}
	assert.Equal(t, ip, a.AAAA.String())
}
func checkCNAME(t *testing.T, rr dns.RR, name string, target string) {
	assert.Equal(t, name, rr.Header().Name)
	assert.Equal(t, dns.TypeCNAME, rr.Header().Rrtype)

	a, ok := rr.(*dns.CNAME)
	if !ok {
		t.Errorf("Found resource record is not a CNAME record: got %s", reflect.TypeOf(rr))
		return
	}
	assert.Equal(t, target, a.Target)
}
