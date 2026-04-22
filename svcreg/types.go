// Package svcreg provides service registry functionality using Consul.
//
// Deprecated: This package is no longer used and will be removed in a future version.
package svcreg

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/swayrider/swlib/logger"
)

// ErrServiceNotFound is returned when no matching service is found in the registry.
var (
	ErrServiceNotFound = errors.New("service not found")
)

// ServiceEntry represents a service instance from the Consul registry.
//
// Deprecated: Use environment-based service discovery instead.
type ServiceEntry struct {
	ServiceName  string
	InstanceName string
	Address      string
	Port         int
	MetaData     Meta
	Tags         []string
}

// ServiceMetaKey is an alias for metadata key strings.
type ServiceMetaKey = string

// ServiceHost returns the host address from metadata (internal or external).
func (e ServiceEntry) ServiceHost(external bool) string {
	key := "internal_host"
	if external {
		key = "external_host"
	}
	return e.MetaData[key]
}

// ServicePort returns the port from metadata (internal or external).
func (e ServiceEntry) ServicePort(external bool) int {
	key := "internal_port"
	if external {
		key = "external_port"
	}
	tmp, err := strconv.Atoi(e.MetaData[key])
	if err != nil {
		return 0
	}
	return tmp
}

// ServiceQuery represents a query for finding services by name and tags.
type ServiceQuery string

// NewServiceQuery creates a query for a service with optional tag filters.
func NewServiceQuery(name string, tags ...string) ServiceQuery {
	tagDesc := strings.Join(tags, "&")
	return ServiceQuery(fmt.Sprintf("%s:%s", name, tagDesc))
}

// Name returns the service name from the query.
func (s ServiceQuery) Name() string {
	return strings.Split(string(s), ":")[0]
}

// Tags returns the tag filters from the query.
func (s ServiceQuery) Tags() []string {
	tags := strings.Split(strings.Split(string(s), ":")[1], "&")
	if len(tags) == 1 && tags[0] == "" {
		return nil
	}
	return tags
}

// Meta is an alias for service metadata.
type Meta = map[string]string

// Resolver maintains a local cache of service instances from Consul.
//
// Deprecated: Use environment-based service discovery instead.
type Resolver struct {
	addresses []string
	insecure  bool
	services  []string
	instances map[string][]ServiceEntry
	mutex     sync.Mutex
	interval  time.Duration
	stopChan  chan struct{}
	logger    *log.Logger
}

// MatchTags returns true if the entry has all specified tags.
func (e ServiceEntry) MatchTags(tags []string) bool {
	// all elements in tags must match
	for _, tag := range tags {
		if !e.TagsContains(tag) {
			return false
		}
	}
	return true
}

// TagsContains returns true if the entry has the specified tag.
func (e ServiceEntry) TagsContains(tag string) bool {
	return slices.Contains(e.Tags, tag)
}
