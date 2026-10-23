//go:build linux || freebsd
// +build linux freebsd

package config

import (
	"github.com/docker/docker/opts"
	units "github.com/docker/go-units"
)

// Config defines the configuration of a docker daemon.
// It includes json tags to deserialize configuration from a file
// using the same names that the flags in the command line uses.
type Config struct {
	CommonConfig

	// These fields are common to all unix platforms.
	CommonUnixConfig

	// Fields below here are platform specific.
	RemappedRoot    string                   `json:"userns-remap,omitempty"`
	Ulimits         map[string]*units.Ulimit `json:"default-ulimits,omitempty"`
	OOMScoreAdjust  int                      `json:"oom-score-adjust,omitempty"`
	ShmSize         opts.MemBytes            `json:"default-shm-size,omitempty"`
	NoNewPrivileges bool                     `json:"no-new-privileges,omitempty"`
}

// BridgeConfig stores all the bridge driver specific
// configuration.
type BridgeConfig struct {
	commonBridgeConfig

	// These fields are common to all unix platforms.
	commonUnixBridgeConfig

	// Fields below here are platform specific.
	EnableIPv6          bool   `json:"ipv6,omitempty"`
	EnableIPTables      bool   `json:"iptables,omitempty"`
	EnableIPForward     bool   `json:"ip-forward,omitempty"`
	EnableIPMasq        bool   `json:"ip-masq,omitempty"`
	EnableUserlandProxy bool   `json:"userland-proxy,omitempty"`
	FixedCIDRv6         string `json:"fixed-cidr-v6,omitempty"`
}
