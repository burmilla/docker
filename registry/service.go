package registry

import (
	"crypto/tls"
	"net/url"
	"sync"

	"github.com/docker/distribution/reference"
	registrytypes "github.com/docker/docker/api/types/registry"
)

// Service is the interface defining what a registry service should implement.
type Service interface {
	LookupPullEndpoints(hostname string) (endpoints []APIEndpoint, err error)
	ResolveRepository(name reference.Named) (*RepositoryInfo, error)
	ServiceConfig() *registrytypes.ServiceConfig
	TLSConfig(hostname string) (*tls.Config, error)
	LoadMirrors([]string) error
	LoadInsecureRegistries([]string) error
}

// DefaultService is a registry service. It tracks configuration data such as a list
// of mirrors.
type DefaultService struct {
	config *serviceConfig
	mu     sync.Mutex
}

// NewService returns a new instance of DefaultService ready to be
// installed into an engine.
func NewService(options ServiceOptions) *DefaultService {
	return &DefaultService{
		config: newServiceConfig(options),
	}
}

// ServiceConfig returns the public registry service configuration.
func (s *DefaultService) ServiceConfig() *registrytypes.ServiceConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	servConfig := registrytypes.ServiceConfig{
		InsecureRegistryCIDRs: make([]*(registrytypes.NetIPNet), 0),
		IndexConfigs:          make(map[string]*(registrytypes.IndexInfo)),
		Mirrors:               make([]string, 0),
	}

	// construct a new ServiceConfig which will not retrieve s.Config directly,
	// and look up items in s.config with mu locked
	servConfig.InsecureRegistryCIDRs = append(servConfig.InsecureRegistryCIDRs, s.config.ServiceConfig.InsecureRegistryCIDRs...)

	for key, value := range s.config.ServiceConfig.IndexConfigs {
		servConfig.IndexConfigs[key] = value
	}

	servConfig.Mirrors = append(servConfig.Mirrors, s.config.ServiceConfig.Mirrors...)

	return &servConfig
}

// LoadMirrors loads registry mirrors for Service
func (s *DefaultService) LoadMirrors(mirrors []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.config.LoadMirrors(mirrors)
}

// LoadInsecureRegistries loads insecure registries for Service
func (s *DefaultService) LoadInsecureRegistries(registries []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.config.LoadInsecureRegistries(registries)
}

// ResolveRepository splits a repository name into its components
// and configuration of the associated registry.
func (s *DefaultService) ResolveRepository(name reference.Named) (*RepositoryInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return newRepositoryInfo(s.config, name)
}

// APIEndpoint represents a remote API endpoint
type APIEndpoint struct {
	Mirror       bool
	URL          *url.URL
	Version      APIVersion
	Official     bool
	TrimHostname bool
	TLSConfig    *tls.Config
}

// TLSConfig constructs a client TLS configuration based on server defaults
func (s *DefaultService) TLSConfig(hostname string) (*tls.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return newTLSConfig(hostname, isSecureIndex(s.config, hostname))
}

// tlsConfig constructs a client TLS configuration based on server defaults
func (s *DefaultService) tlsConfig(hostname string) (*tls.Config, error) {
	return newTLSConfig(hostname, isSecureIndex(s.config, hostname))
}

func (s *DefaultService) tlsConfigForMirror(mirrorURL *url.URL) (*tls.Config, error) {
	return s.tlsConfig(mirrorURL.Host)
}

// LookupPullEndpoints creates a list of endpoints to try to pull from, in order of preference.
// It gives preference to mirrors over the actual registry, and HTTPS over
// plain HTTP.
func (s *DefaultService) LookupPullEndpoints(hostname string) (endpoints []APIEndpoint, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lookupV2Endpoints(hostname)
}
