//go:build !windows && !solaris
// +build !windows,!solaris

package main

import (
	"fmt"
	"net"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/cmd/dockerd/hack"
	"github.com/docker/docker/daemon"
	"github.com/docker/docker/libcontainerd"
)

const defaultDaemonConfigFile = "/etc/docker/system-docker.json"

// ignoreHangup keeps a SIGHUP (which older releases used to reload the
// config file) from terminating the daemon.
func ignoreHangup() {
	signal.Ignore(syscall.SIGHUP)
}

// setDefaultUmask sets the umask to 0022 to avoid problems
// caused by custom umask
func setDefaultUmask() error {
	desiredUmask := 0022
	syscall.Umask(desiredUmask)
	if umask := syscall.Umask(desiredUmask); umask != desiredUmask {
		return fmt.Errorf("failed to set umask: expected %#o, got %#o", desiredUmask, umask)
	}

	return nil
}

func (cli *DaemonCli) getPlatformRemoteOptions() []libcontainerd.RemoteOption {
	opts := []libcontainerd.RemoteOption{
		libcontainerd.WithDebugLog(cli.Config.Debug),
		libcontainerd.WithOOMScore(cli.Config.OOMScoreAdjust),
	}
	if cli.Config.ContainerdAddr != "" {
		opts = append(opts, libcontainerd.WithRemoteAddr(cli.Config.ContainerdAddr))
	} else {
		opts = append(opts, libcontainerd.WithStartDaemon(true))
	}
	if cli.Config.LiveRestoreEnabled {
		opts = append(opts, libcontainerd.WithLiveRestore(true))
	}
	opts = append(opts, libcontainerd.WithRuntimePath(daemon.DefaultRuntimeBinary))
	return opts
}

// getLibcontainerdRoot gets the root directory for libcontainerd/containerd to
// store their state.
func (cli *DaemonCli) getLibcontainerdRoot() string {
	return filepath.Join(cli.Config.ExecRoot, "libcontainerd")
}

// notifyShutdown is called after the daemon shuts down but before the process exits.
func notifyShutdown(err error) {
}

func wrapListeners(ls []net.Listener) []net.Listener {
	ls[0] = &hack.MalformedHostHeaderOverride{Listener: ls[0]}
	return ls
}
