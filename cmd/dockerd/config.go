package main

import (
	"runtime"

	"github.com/docker/docker/daemon/config"
	"github.com/docker/docker/opts"
	"github.com/spf13/pflag"
)

const (
	// defaultShutdownTimeout is the default shutdown timeout for the daemon
	defaultShutdownTimeout = 15
)

// installCommonConfigFlags adds flags to the pflag.FlagSet to configure the daemon
func installCommonConfigFlags(conf *config.Config, flags *pflag.FlagSet) {
	var maxConcurrentDownloads int

	conf.ServiceOptions.InstallCliFlags(flags)

	flags.Var(opts.NewNamedListOptsRef("storage-opts", &conf.GraphOptions, nil), "storage-opt", "Storage driver options")
	flags.StringVarP(&conf.Pidfile, "pidfile", "p", defaultPidFile, "Path to use for daemon PID file")
	flags.StringVar(&conf.Root, "data-root", defaultDataRoot, "Root directory of persistent Docker state")

	flags.BoolVarP(&conf.AutoRestart, "restart", "r", true, "--restart on the daemon has been deprecated in favor of --restart policies on docker run")
	flags.MarkDeprecated("restart", "Please use a restart policy on docker run")

	// Windows doesn't support setting the storage driver - there is no choice as to which ones to use.
	if runtime.GOOS != "windows" {
		flags.StringVarP(&conf.GraphDriver, "storage-driver", "s", "", "Storage driver to use")
	}

	flags.IntVar(&conf.Mtu, "mtu", 0, "Set the containers network MTU")
	flags.BoolVar(&conf.RawLogs, "raw-logs", false, "Full timestamps without ANSI coloring")
	flags.Var(opts.NewListOptsRef(&conf.DNS, opts.ValidateIPAddress), "dns", "DNS server to use")
	flags.Var(opts.NewNamedListOptsRef("dns-opts", &conf.DNSOptions, nil), "dns-opt", "DNS options to use")
	flags.Var(opts.NewListOptsRef(&conf.DNSSearch, opts.ValidateDNSSearch), "dns-search", "DNS search domains to use")
	flags.Var(opts.NewNamedListOptsRef("labels", &conf.Labels, opts.ValidateLabel), "label", "Set key=value labels to the daemon")
	flags.StringVar(&conf.LogConfig.Type, "log-driver", "json-file", "Default driver for container logs")
	flags.Var(opts.NewNamedMapOpts("log-opts", conf.LogConfig.Config, nil), "log-opt", "Default log driver options for containers")
	flags.IntVar(&maxConcurrentDownloads, "max-concurrent-downloads", config.DefaultMaxConcurrentDownloads, "Set the max concurrent downloads for each pull")
	flags.IntVar(&conf.ShutdownTimeout, "shutdown-timeout", defaultShutdownTimeout, "Set the default shutdown timeout")

	conf.MaxConcurrentDownloads = &maxConcurrentDownloads
}
