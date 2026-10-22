//go:build !windows && !solaris
// +build !windows,!solaris

package runconfig

import (
	"fmt"
	"runtime"

	"github.com/docker/docker/api/types/container"
)

// DefaultDaemonNetworkMode returns the default network stack the daemon should
// use.
func DefaultDaemonNetworkMode() container.NetworkMode {
	return container.NetworkMode("bridge")
}

// IsPreDefinedNetwork indicates if a network is predefined by the daemon
func IsPreDefinedNetwork(network string) bool {
	n := container.NetworkMode(network)
	return n.IsBridge() || n.IsHost() || n.IsNone() || n.IsDefault()
}

// validateNetMode ensures that the various combinations of requested
// network settings are valid.
func validateNetMode(c *container.Config, hc *container.HostConfig) error {
	// We may not be passed a host config, such as in the case of docker commit
	if hc == nil {
		return nil
	}

	err := validateNetContainerMode(c, hc)
	if err != nil {
		return err
	}

	if hc.UTSMode.IsHost() && c.Hostname != "" {
		return ErrConflictUTSHostname
	}

	return nil
}

// validateIsolation performs platform specific validation of
// isolation in the hostconfig structure. Linux only supports "default"
// which is LXC container isolation
func validateIsolation(hc *container.HostConfig) error {
	// We may not be passed a host config, such as in the case of docker commit
	if hc == nil {
		return nil
	}
	if !hc.Isolation.IsValid() {
		return fmt.Errorf("Invalid isolation: %q - %s only supports 'default'", hc.Isolation, runtime.GOOS)
	}
	return nil
}

// validatePrivileged performs platform specific validation of the Privileged setting
func validatePrivileged(hc *container.HostConfig) error {
	return nil
}

// validateReadonlyRootfs performs platform specific validation of the ReadonlyRootfs setting
func validateReadonlyRootfs(hc *container.HostConfig) error {
	return nil
}
