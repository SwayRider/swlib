// Package svcreg provides service registry functionality using Consul.
//
// Deprecated: This package is no longer used and will be removed in a future version.
package svcreg

import (
	"math/rand"
	"time"

	"github.com/hashicorp/consul/api"
	log "github.com/swayrider/swlib/logger"
)

// NewResolver creates a new Consul-based service resolver.
//
// Deprecated: Use environment-based service discovery instead.
func NewResolver(
	consulAddrs []string,
	insecure bool,
	services []string,
	interval time.Duration,
	l *log.Logger) *Resolver {
	return &Resolver{
		addresses: consulAddrs,
		insecure:  insecure,
		services:  services,
		instances: make(map[string][]ServiceEntry),
		interval:  interval,
		stopChan:  make(chan struct{}),
		logger: l.Derive(
			log.WithComponent("svcreg.Resolver"),
			log.WithFunction("NewResolver")),
	}
}

func (r *Resolver) syncLoop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.sync()
		case <-r.stopChan:
			return
		}
	}
}

func (r *Resolver) sync() {
	var lastErr error

	for _, addr := range r.addresses {
		config := api.DefaultConfig()
		config.Address = addr
		config.TLSConfig.InsecureSkipVerify = r.insecure

		client, err := api.NewClient(config)
		if err != nil {
			lastErr = err
			continue
		}

		lastErr = r.trySyncClient(client)
		if lastErr == nil {
			return
		}
	}

	if lastErr != nil {
		r.logger.Derive(log.WithFunction("sync")).Errorf(
			"All consul addresses failed: %v", lastErr)
	}
}

func (r *Resolver) trySyncClient(client *api.Client) error {
	lg := r.logger.Derive(log.WithFunction("trySyncClient"))

	_, err := client.Agent().Self()
	if err != nil {
		return err
	}

	for _, svc := range r.services {
		entries, _, err := client.Health().Service(svc, "", true, nil)
		if err != nil {
			lg.Warnf("failed to get service %s: %v", svc, err)
			continue
		}

		var healthy []ServiceEntry
		for _, entry := range entries {
			addr := entry.Service.Address
			if addr == "" {
				addr = entry.Node.Address
			}

			healthy = append(healthy, ServiceEntry{
				ServiceName:  svc,
				InstanceName: entry.Service.ID,
				Address:      addr,
				Port:         entry.Service.Port,
				MetaData:     entry.Service.Meta,
				Tags:         entry.Service.Tags,
			})
		}

		r.mutex.Lock()
		r.instances[svc] = healthy
		r.mutex.Unlock()
	}
	return nil
}

// Start begins the background synchronization loop.
func (r *Resolver) Start() {
	// Initial sync
	r.sync()
	go r.syncLoop()
}

// Stop terminates the background synchronization loop.
func (r *Resolver) Stop() {
	close(r.stopChan)
}

// Get returns a random matching service instance for the given query.
// Returns ErrServiceNotFound if no matching services are available.
func (r *Resolver) Get(query ServiceQuery) (ServiceEntry, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	serviceName := query.Name()
	tags := query.Tags()

	cands := make([]ServiceEntry, 0)
	if serviceName == "" {
		for _, entries := range r.instances {
			for _, entry := range entries {
				if entry.MatchTags(tags) {
					cands = append(cands, entry)
				}
			}
		}
	} else {
		entries := r.instances[serviceName]
		for _, entry := range entries {
			if entry.MatchTags(tags) {
				cands = append(cands, entry)
			}
		}
	}

	if len(cands) == 0 {
		return ServiceEntry{}, ErrServiceNotFound
	}

	return cands[rand.Intn(len(cands))], nil
}
