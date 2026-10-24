//go:build solaris || linux || freebsd
// +build solaris linux freebsd

package config

import (
	"net"
)

// CommonUnixConfig defines configuration of a docker daemon that is
// common across Unix platforms.
type CommonUnixConfig struct {
	ExecRoot       string `json:"exec-root,omitempty"`
	ContainerdAddr string `json:"containerd,omitempty"`
}

type commonUnixBridgeConfig struct {
	DefaultIP                   net.IP `json:"ip,omitempty"`
	IP                          string `json:"bip,omitempty"`
	DefaultGatewayIPv4          net.IP `json:"default-gateway,omitempty"`
	DefaultGatewayIPv6          net.IP `json:"default-gateway-v6,omitempty"`
	InterContainerCommunication bool   `json:"icc,omitempty"`
}

// GetExecRoot returns the user configured Exec-root
func (conf *Config) GetExecRoot() string {
	return conf.ExecRoot
}
