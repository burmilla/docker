package daemon

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/config"
	"github.com/docker/docker/libcontainerd"
)

// Reload reads configuration changes and modifies the
// daemon according to those changes.
// These are the settings that Reload changes:
// - Platform runtime
// - Daemon debug log level
// - Daemon max concurrent downloads
// - Daemon shutdown timeout (in seconds)
// - Daemon labels
// - Insecure registries
// - Registry mirrors
// - Daemon live restore
func (daemon *Daemon) Reload(conf *config.Config) (err error) {
	daemon.configStore.Lock()
	attributes := map[string]string{}

	defer func() {
		// we're unlocking here, because
		// LogDaemonEventWithAttributes() -> SystemInfo() -> GetAllRuntimes()
		// holds that lock too.
		daemon.configStore.Unlock()
		if err == nil {
			daemon.LogDaemonEventWithAttributes("reload", attributes)
		}
	}()

	daemon.reloadPlatform(conf, attributes)
	daemon.reloadDebug(conf, attributes)
	daemon.reloadMaxConcurrentDowloadsAndUploads(conf, attributes)
	daemon.reloadShutdownTimeout(conf, attributes)

	if err := daemon.reloadLabels(conf, attributes); err != nil {
		return err
	}
	if err := daemon.reloadInsecureRegistries(conf, attributes); err != nil {
		return err
	}
	if err := daemon.reloadRegistryMirrors(conf, attributes); err != nil {
		return err
	}
	if err := daemon.reloadLiveRestore(conf, attributes); err != nil {
		return err
	}
	return nil
}

// reloadDebug updates configuration with Debug option
// and updates the passed attributes
func (daemon *Daemon) reloadDebug(conf *config.Config, attributes map[string]string) {
	// update corresponding configuration
	if conf.IsValueSet("debug") {
		daemon.configStore.Debug = conf.Debug
	}
	// prepare reload event attributes with updatable configurations
	attributes["debug"] = fmt.Sprintf("%t", daemon.configStore.Debug)
}

// reloadMaxConcurrentDowloadsAndUploads updates configuration with max concurrent
// download and upload options and updates the passed attributes
func (daemon *Daemon) reloadMaxConcurrentDowloadsAndUploads(conf *config.Config, attributes map[string]string) {
	// If no value is set for max-concurrent-downloads we assume it is the default value
	// We always "reset" as the cost is lightweight and easy to maintain.
	if conf.IsValueSet("max-concurrent-downloads") && conf.MaxConcurrentDownloads != nil {
		*daemon.configStore.MaxConcurrentDownloads = *conf.MaxConcurrentDownloads
	} else {
		maxConcurrentDownloads := config.DefaultMaxConcurrentDownloads
		daemon.configStore.MaxConcurrentDownloads = &maxConcurrentDownloads
	}
	logrus.Debugf("Reset Max Concurrent Downloads: %d", *daemon.configStore.MaxConcurrentDownloads)
	if daemon.downloadManager != nil {
		daemon.downloadManager.SetConcurrency(*daemon.configStore.MaxConcurrentDownloads)
	}

	// prepare reload event attributes with updatable configurations
	attributes["max-concurrent-downloads"] = fmt.Sprintf("%d", *daemon.configStore.MaxConcurrentDownloads)
}

// reloadShutdownTimeout updates configuration with daemon shutdown timeout option
// and updates the passed attributes
func (daemon *Daemon) reloadShutdownTimeout(conf *config.Config, attributes map[string]string) {
	// update corresponding configuration
	if conf.IsValueSet("shutdown-timeout") {
		daemon.configStore.ShutdownTimeout = conf.ShutdownTimeout
		logrus.Debugf("Reset Shutdown Timeout: %d", daemon.configStore.ShutdownTimeout)
	}

	// prepare reload event attributes with updatable configurations
	attributes["shutdown-timeout"] = fmt.Sprintf("%d", daemon.configStore.ShutdownTimeout)
}

// reloadLabels updates configuration with engine labels
// and updates the passed attributes
func (daemon *Daemon) reloadLabels(conf *config.Config, attributes map[string]string) error {
	// update corresponding configuration
	if conf.IsValueSet("labels") {
		daemon.configStore.Labels = conf.Labels
	}

	// prepare reload event attributes with updatable configurations
	if daemon.configStore.Labels != nil {
		labels, err := json.Marshal(daemon.configStore.Labels)
		if err != nil {
			return err
		}
		attributes["labels"] = string(labels)
	} else {
		attributes["labels"] = "[]"
	}

	return nil
}

// reloadInsecureRegistries updates configuration with insecure registry option
// and updates the passed attributes
func (daemon *Daemon) reloadInsecureRegistries(conf *config.Config, attributes map[string]string) error {
	// update corresponding configuration
	if conf.IsValueSet("insecure-registries") {
		daemon.configStore.InsecureRegistries = conf.InsecureRegistries
		if err := daemon.RegistryService.LoadInsecureRegistries(conf.InsecureRegistries); err != nil {
			return err
		}
	}

	// prepare reload event attributes with updatable configurations
	if daemon.configStore.InsecureRegistries != nil {
		insecureRegistries, err := json.Marshal(daemon.configStore.InsecureRegistries)
		if err != nil {
			return err
		}
		attributes["insecure-registries"] = string(insecureRegistries)
	} else {
		attributes["insecure-registries"] = "[]"
	}

	return nil
}

// reloadRegistryMirrors updates configuration with registry mirror options
// and updates the passed attributes
func (daemon *Daemon) reloadRegistryMirrors(conf *config.Config, attributes map[string]string) error {
	// update corresponding configuration
	if conf.IsValueSet("registry-mirrors") {
		daemon.configStore.Mirrors = conf.Mirrors
		if err := daemon.RegistryService.LoadMirrors(conf.Mirrors); err != nil {
			return err
		}
	}

	// prepare reload event attributes with updatable configurations
	if daemon.configStore.Mirrors != nil {
		mirrors, err := json.Marshal(daemon.configStore.Mirrors)
		if err != nil {
			return err
		}
		attributes["registry-mirrors"] = string(mirrors)
	} else {
		attributes["registry-mirrors"] = "[]"
	}

	return nil
}

// reloadLiveRestore updates configuration with live retore option
// and updates the passed attributes
func (daemon *Daemon) reloadLiveRestore(conf *config.Config, attributes map[string]string) error {
	// update corresponding configuration
	if conf.IsValueSet("live-restore") {
		daemon.configStore.LiveRestoreEnabled = conf.LiveRestoreEnabled
		if err := daemon.containerdRemote.UpdateOptions(libcontainerd.WithLiveRestore(conf.LiveRestoreEnabled)); err != nil {
			return err
		}
	}

	// prepare reload event attributes with updatable configurations
	attributes["live-restore"] = fmt.Sprintf("%t", daemon.configStore.LiveRestoreEnabled)
	return nil
}
