//go:build linux || freebsd
// +build linux freebsd

package daemon

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/container"
	"github.com/docker/docker/daemon/links"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/runconfig"
	"github.com/docker/libnetwork"
	"github.com/pkg/errors"
)

func (daemon *Daemon) setupLinkedContainers(container *container.Container) ([]string, error) {
	var env []string
	children := daemon.children(container)

	bridgeSettings := container.NetworkSettings.Networks[runconfig.DefaultDaemonNetworkMode().NetworkName()]
	if bridgeSettings == nil || bridgeSettings.EndpointSettings == nil {
		return nil, nil
	}

	for linkAlias, child := range children {
		if !child.IsRunning() {
			return nil, fmt.Errorf("Cannot link to a non running container: %s AS %s", child.Name, linkAlias)
		}

		childBridgeSettings := child.NetworkSettings.Networks[runconfig.DefaultDaemonNetworkMode().NetworkName()]
		if childBridgeSettings == nil || childBridgeSettings.EndpointSettings == nil {
			return nil, fmt.Errorf("container %s not attached to default bridge network", child.ID)
		}

		link := links.NewLink(
			bridgeSettings.IPAddress,
			childBridgeSettings.IPAddress,
			linkAlias,
			child.Config.Env,
			child.Config.ExposedPorts,
		)

		env = append(env, link.ToEnv()...)
	}

	return env, nil
}

func (daemon *Daemon) getIpcContainer(container *container.Container) (*container.Container, error) {
	containerID := container.HostConfig.IpcMode.Container()
	container, err := daemon.GetContainer(containerID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot join IPC of a non running container: %s", container.ID)
	}
	return container, daemon.checkContainer(container, containerIsRunning, containerIsNotRestarting)
}

func (daemon *Daemon) getPidContainer(container *container.Container) (*container.Container, error) {
	containerID := container.HostConfig.PidMode.Container()
	container, err := daemon.GetContainer(containerID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot join PID of a non running container: %s", container.ID)
	}
	return container, daemon.checkContainer(container, containerIsRunning, containerIsNotRestarting)
}

func containerIsRunning(c *container.Container) error {
	if !c.IsRunning() {
		return errors.Errorf("container %s is not running", c.ID)
	}
	return nil
}

func containerIsNotRestarting(c *container.Container) error {
	if c.IsRestarting() {
		return errContainerIsRestarting(c.ID)
	}
	return nil
}

func (daemon *Daemon) setupIpcDirs(c *container.Container) error {
	var err error

	c.ShmPath, err = c.ShmResourcePath()
	if err != nil {
		return err
	}

	if c.HostConfig.IpcMode.IsContainer() {
		ic, err := daemon.getIpcContainer(c)
		if err != nil {
			return err
		}
		c.ShmPath = ic.ShmPath
	} else if c.HostConfig.IpcMode.IsHost() {
		if _, err := os.Stat("/dev/shm"); err != nil {
			return fmt.Errorf("/dev/shm is not mounted, but must be for --ipc=host")
		}
		c.ShmPath = "/dev/shm"
	} else {
		rootIDs := daemon.idMappings.RootPair()
		if !c.HasMountFor("/dev/shm") {
			shmPath, err := c.ShmResourcePath()
			if err != nil {
				return err
			}

			if err := idtools.MkdirAllAndChown(shmPath, 0700, rootIDs); err != nil {
				return err
			}

			shmSize := int64(daemon.configStore.ShmSize)
			if c.HostConfig.ShmSize != 0 {
				shmSize = c.HostConfig.ShmSize
			}
			shmproperty := "mode=1777,size=" + strconv.FormatInt(shmSize, 10)
			if err := syscall.Mount("shm", shmPath, "tmpfs", uintptr(syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV), shmproperty); err != nil {
				return fmt.Errorf("mounting shm tmpfs: %s", err)
			}
			if err := os.Chown(shmPath, rootIDs.UID, rootIDs.GID); err != nil {
				return err
			}
		}

	}

	return nil
}

func killProcessDirectly(cntr *container.Container) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Block until the container to stops or timeout.
	status := <-cntr.Wait(ctx, container.WaitConditionNotRunning)
	if status.Err() != nil {
		// Ensure that we don't kill ourselves
		if pid := cntr.GetPID(); pid != 0 {
			logrus.Infof("Container %s failed to exit within 10 seconds of kill - trying direct SIGKILL", stringid.TruncateID(cntr.ID))
			if err := syscall.Kill(pid, 9); err != nil {
				if err != syscall.ESRCH {
					return err
				}
				e := errNoSuchProcess{pid, 9}
				logrus.Debug(e)
				return e
			}
		}
	}
	return nil
}

func detachMounted(path string) error {
	return syscall.Unmount(path, syscall.MNT_DETACH)
}

func isLinkable(child *container.Container) bool {
	// A container is linkable only if it belongs to the default network
	_, ok := child.NetworkSettings.Networks[runconfig.DefaultDaemonNetworkMode().NetworkName()]
	return ok
}

func enableIPOnPredefinedNetwork() bool {
	return false
}

func (daemon *Daemon) isNetworkHotPluggable() bool {
	return true
}

func setupPathsAndSandboxOptions(container *container.Container, sboxOptions *[]libnetwork.SandboxOption) error {
	var err error

	container.HostsPath, err = container.GetRootResourcePath("hosts")
	if err != nil {
		return err
	}
	*sboxOptions = append(*sboxOptions, libnetwork.OptionHostsPath(container.HostsPath))

	container.ResolvConfPath, err = container.GetRootResourcePath("resolv.conf")
	if err != nil {
		return err
	}
	*sboxOptions = append(*sboxOptions, libnetwork.OptionResolvConfPath(container.ResolvConfPath))
	return nil
}

func initializeNetworkingPaths(container *container.Container, nc *container.Container) {
	container.HostnamePath = nc.HostnamePath
	container.HostsPath = nc.HostsPath
	container.ResolvConfPath = nc.ResolvConfPath
}
