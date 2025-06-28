package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/miekg/dns"
)

func createCacheKey(msg *dns.Msg) string {
	if len(msg.Question) > 0 {
		return msg.Question[0].Name + "|" + strconv.Itoa(int(msg.Question[0].Qtype))
	}
	return ""
}

type CachedResponse struct {
	Response *dns.Msg
	CachedAt time.Time
}

type DNSService struct {
	upstreamServerSelector *DNSServerSelector
	filteringService       *FilteringService
	verbose                bool
	cache                  *ttlcache.Cache
}

func NewDNSService(serverSelector *DNSServerSelector, filteringService *FilteringService,
	configService *ApplicationConfigService) *DNSService {
	return &DNSService{
		upstreamServerSelector: serverSelector,
		filteringService:       filteringService,
		cache:                  ttlcache.NewCache(),
	}
}

func (dnsService *DNSService) ProcessQuery(ctx context.Context, msg *dns.Msg, configId string) (*dns.Msg, error) {
	startTime := time.Now()
	domain := msg.Question[0].Name
	blocked, err := dnsService.filteringService.IsDomainBlocked(ctx, configId, domain)
	if err != nil {
		msg.Rcode = dns.RcodeNameError //send NXDOMAIN
		elapsedTime := time.Since(startTime)
		slog.Error("Error during processig query", "rcode", msg.Rcode, "elapsedTime", elapsedTime)
		return msg, nil
	}
	if blocked {
		//rr := new(dns.A)
		//rr.Hdr = dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
		//rr.A = net.ParseIP("0.0.0.0")
		//msg.Answer = append(msg.Answer, rr)
		msg.Rcode = dns.RcodeNameError //send NXDOMAIN
		elapsedTime := time.Since(startTime)
		slog.Debug("Replying back", "domain", domain, "rcode", msg.Rcode, "elapsedTime", elapsedTime)
		return msg, nil
	}

	//check in cache
	key := createCacheKey(msg)
	slog.Debug("checking cache for", "key", key)
	cacheResponse, ok := dnsService.cache.Get(key)
	if ok {
		elapsedTime := time.Since(startTime)
		slog.Debug("replying back from cache", "elapsedTime", elapsedTime)
		return cacheResponse.(CachedResponse).Response, nil
	}

	slog.Debug("querying upstream domain", "domainName", domain)
	response, err := dnsService.QueryUpstream(msg)
	if err != nil {
		return nil, err
	}

	//add to cache
	leastTTL := uint32(math.MaxUint32)
	for _, v := range response.Answer {
		leastTTL = min(leastTTL, v.Header().Ttl)
	}

	cacheRes := CachedResponse{
		Response: response,
		CachedAt: time.Now(),
	}
	slog.Debug("adding to cache", "key", key, "ttl", leastTTL)
	dnsService.cache.SetWithTTL(key, cacheRes, time.Duration(leastTTL)*time.Second)

	elapsedTime := time.Since(startTime)
	slog.Debug("Replying back", "domain", domain, "rcode", response.Rcode, "elapsedTime", elapsedTime)
	return response, nil
}

func (dnsService *DNSService) QueryUpstream(msg *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	c.Net = "udp"

	// Set a timeout to prevent hanging
	c.Timeout = 5 * time.Second

	upstreamServer := dnsService.upstreamServerSelector.GetNext()
	// Forward to upstream DNS server
	response, _, err := c.Exchange(msg, fmt.Sprintf("%s:%d", upstreamServer, 53))
	if err != nil {
		if dnsService.verbose {
			log.Printf("Error querying upstream DNS: %v", err)
		}

		// If UDP fails, try TCP
		c.Net = "tcp"
		response, _, err = c.Exchange(msg, fmt.Sprintf("%s:%d", upstreamServer, 53))
		if err != nil {
			return nil, err
		}
	}

	if dnsService.verbose && response != nil {
		log.Printf("Got response with code: %s, %d answers", dns.RcodeToString[response.Rcode], len(response.Answer))
	}
	slog.Debug("request processed by", "upstreamServer", upstreamServer)
	return response, nil
}

// DNSServerSelector manages the round-robin selection of DNS servers
type DNSServerSelector struct {
	servers []string
	current uint32
	mu      sync.RWMutex
}

func NewDNSServerSelector(servers []string) *DNSServerSelector {
	return &DNSServerSelector{
		servers: servers,
	}
}

// GetNext returns the next DNS server in round-robin fashion
func (s *DNSServerSelector) GetNext() string {
	// Get the current length of servers slice
	s.mu.RLock()
	length := uint32(len(s.servers))
	s.mu.RUnlock()

	// Read current value first, then increment
	current := atomic.LoadUint32(&s.current)
	atomic.AddUint32(&s.current, 1)
	index := current % length

	// Get the server at calculated index
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.servers[index]
}

// AddServer adds a new DNS server to the pool
func (s *DNSServerSelector) AddServer(server string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.servers = append(s.servers, server)
}

// RemoveServer removes a DNS server from the pool
func (s *DNSServerSelector) RemoveServer(server string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, srv := range s.servers {
		if srv == server {
			// Remove the server by slicing
			s.servers = append(s.servers[:i], s.servers[i+1:]...)
			// Ensure we still have at least one server
			if len(s.servers) == 0 {
				panic("cannot remove last DNS server")
			}
			return true
		}
	}
	return false
}

// GetServers returns a copy of the current server list
func (s *DNSServerSelector) GetServers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servers := make([]string, len(s.servers))
	copy(servers, s.servers)
	return servers
}
