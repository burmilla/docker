package daemon

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork"
)

// ContainerRename changes the name of a container, using the oldName
// to find the container. An error is returned if newName is already
// reserved.
func (daemon *Daemon) ContainerRename(oldName, newName string) error {
	var (
		sid string
		sb  libnetwork.Sandbox
	)

	if oldName == "" || newName == "" {
		return errors.New("Neither old nor new names may be empty")
	}

	if newName[0] != '/' {
		newName = "/" + newName
	}

	container, err := daemon.GetContainer(oldName)
	if err != nil {
		return err
	}

	oldName = container.Name
	oldIsAnonymousEndpoint := container.NetworkSettings.IsAnonymousEndpoint

	if oldName == newName {
		return errors.New("Renaming a container with the same name as its current name")
	}

	container.Lock()
	defer container.Unlock()

	if newName, err = daemon.reserveName(container.ID, newName); err != nil {
		return fmt.Errorf("Error when allocating new name: %v", err)
	}

	container.Name = newName
	container.NetworkSettings.IsAnonymousEndpoint = false

	defer func() {
		if err != nil {
			container.Name = oldName
			container.NetworkSettings.IsAnonymousEndpoint = oldIsAnonymousEndpoint
			daemon.reserveName(container.ID, oldName)
			daemon.releaseName(newName)
		}
	}()

	daemon.releaseName(oldName)
	if err = container.CheckpointTo(daemon.containersReplica); err != nil {
		return err
	}

	attributes := map[string]string{
		"oldName": oldName,
	}

	if !container.Running {
		daemon.LogContainerEventWithAttributes(container, "rename", attributes)
		return nil
	}

	defer func() {
		if err != nil {
			container.Name = oldName
			container.NetworkSettings.IsAnonymousEndpoint = oldIsAnonymousEndpoint
			if e := container.CheckpointTo(daemon.containersReplica); e != nil {
				logrus.Errorf("%s: Failed in writing to Disk on rename failure: %v", container.ID, e)
			}
		}
	}()

	sid = container.NetworkSettings.SandboxID
	if sid != "" && daemon.netController != nil {
		sb, err = daemon.netController.SandboxByID(sid)
		if err != nil {
			return err
		}

		err = sb.Rename(strings.TrimPrefix(container.Name, "/"))
		if err != nil {
			return err
		}
	}

	daemon.LogContainerEventWithAttributes(container, "rename", attributes)
	return nil
}
