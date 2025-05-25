package service

import (
	"net"
	"reflect"
	"testing"

	"github.com/miekg/dns"
)

func TestDNSService_QueryUpstream(t *testing.T) {
	type fields struct {
		upstreamServerSelector *DNSServerSelector
		filteringService       *FilteringService
		verbose                bool
	}

	dnsServerSelector := NewDNSServerSelector([]string{"1.1.1.1"})
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	msg.RecursionDesired = true

	ans := new(dns.Msg)
	ans.SetReply(msg)

	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: "google.com", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 132, Rdlength: 4}
	rr.A = net.ParseIP("142.250.70.78").To4()

	ans.Answer = append(ans.Answer, rr)

	type args struct {
		msg *dns.Msg
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *dns.Msg
		wantErr bool
	}{
		{
			name: "test01",
			fields: fields{
				upstreamServerSelector: dnsServerSelector,
				verbose:                true,
			},
			args: args{
				msg: msg,
			},
			want:    ans,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dnsService := &DNSService{
				upstreamServerSelector: tt.fields.upstreamServerSelector,
				filteringService:       tt.fields.filteringService,
				verbose:                tt.fields.verbose,
			}
			got, err := dnsService.QueryUpstream(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("DNSService.QueryUpstream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DNSService.QueryUpstream() = %v, want %v", got, tt.want)
			}
		})
	}
}
