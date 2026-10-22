package daemon

import (
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/backend"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/container"
	"github.com/docker/docker/daemon/network"
	"github.com/docker/go-connections/nat"
)

// ContainerInspect returns low-level information about a
// container. Returns an error if the container cannot be found, or if
// there is an error getting the data.
func (daemon *Daemon) ContainerInspect(name string, size bool) (interface{}, error) {
	return daemon.ContainerInspectCurrent(name, size)
}

// ContainerInspectCurrent returns low-level information about a
// container in a most recent api version.
func (daemon *Daemon) ContainerInspectCurrent(name string, size bool) (*types.ContainerJSON, error) {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	container.Lock()

	base, err := daemon.getInspectData(container)
	if err != nil {
		container.Unlock()
		return nil, err
	}

	apiNetworks := make(map[string]*networktypes.EndpointSettings)
	for name, epConf := range container.NetworkSettings.Networks {
		if epConf.EndpointSettings != nil {
			// We must make a copy of this pointer object otherwise it can race with other operations
			apiNetworks[name] = epConf.EndpointSettings.Copy()
		}
	}

	mountPoints := container.GetMountPoints()
	networkSettings := &types.NetworkSettings{
		NetworkSettingsBase: types.NetworkSettingsBase{
			Bridge:                 container.NetworkSettings.Bridge,
			SandboxID:              container.NetworkSettings.SandboxID,
			HairpinMode:            container.NetworkSettings.HairpinMode,
			LinkLocalIPv6Address:   container.NetworkSettings.LinkLocalIPv6Address,
			LinkLocalIPv6PrefixLen: container.NetworkSettings.LinkLocalIPv6PrefixLen,
			SandboxKey:             container.NetworkSettings.SandboxKey,
			SecondaryIPAddresses:   container.NetworkSettings.SecondaryIPAddresses,
			SecondaryIPv6Addresses: container.NetworkSettings.SecondaryIPv6Addresses,
		},
		DefaultNetworkSettings: daemon.getDefaultNetworkSettings(container.NetworkSettings.Networks),
		Networks:               apiNetworks,
	}

	ports := make(nat.PortMap, len(container.NetworkSettings.Ports))
	for k, pm := range container.NetworkSettings.Ports {
		ports[k] = pm
	}
	networkSettings.NetworkSettingsBase.Ports = ports

	container.Unlock()

	if size {
		sizeRw, sizeRootFs := daemon.getSize(base.ID)
		base.SizeRw = &sizeRw
		base.SizeRootFs = &sizeRootFs
	}

	return &types.ContainerJSON{
		ContainerJSONBase: base,
		Mounts:            mountPoints,
		Config:            container.Config,
		NetworkSettings:   networkSettings,
	}, nil
}

func (daemon *Daemon) getInspectData(container *container.Container) (*types.ContainerJSONBase, error) {
	// make a copy to play with
	hostConfig := *container.HostConfig

	// We merge the Ulimits from hostConfig with daemon default
	daemon.mergeUlimits(&hostConfig)

	containerState := &types.ContainerState{
		Status:     container.State.StateString(),
		Running:    container.State.Running,
		Paused:     container.State.Paused,
		Restarting: container.State.Restarting,
		OOMKilled:  container.State.OOMKilled,
		Dead:       container.State.Dead,
		Pid:        container.State.Pid,
		ExitCode:   container.State.ExitCode(),
		Error:      container.State.ErrorMsg,
		StartedAt:  container.State.StartedAt.Format(time.RFC3339Nano),
		FinishedAt: container.State.FinishedAt.Format(time.RFC3339Nano),
	}

	contJSONBase := &types.ContainerJSONBase{
		ID:           container.ID,
		Created:      container.Created.Format(time.RFC3339Nano),
		Path:         container.Path,
		Args:         container.Args,
		State:        containerState,
		Image:        container.ImageID.String(),
		LogPath:      container.LogPath,
		Name:         container.Name,
		RestartCount: container.RestartCount,
		Driver:       container.Driver,
		Platform:     container.Platform,
		ExecIDs:      container.GetExecIDs(),
		HostConfig:   &hostConfig,
	}

	// Now set any platform-specific fields
	contJSONBase = setPlatformSpecificContainerFields(container, contJSONBase)

	contJSONBase.GraphDriver.Name = container.Driver

	graphDriverData, err := container.RWLayer.Metadata()
	// If container is marked as Dead, the container's graphdriver metadata
	// could have been removed, it will cause error if we try to get the metadata,
	// we can ignore the error if the container is dead.
	if err != nil && !container.Dead {
		return nil, err
	}
	contJSONBase.GraphDriver.Data = graphDriverData

	return contJSONBase, nil
}

// ContainerExecInspect returns low-level information about the exec
// command. An error is returned if the exec cannot be found.
func (daemon *Daemon) ContainerExecInspect(id string) (*backend.ExecInspect, error) {
	e, err := daemon.getExecConfig(id)
	if err != nil {
		return nil, err
	}

	pc := inspectExecProcessConfig(e)

	return &backend.ExecInspect{
		ID:            e.ID,
		Running:       e.Running,
		ExitCode:      e.ExitCode,
		ProcessConfig: pc,
		OpenStdin:     e.OpenStdin,
		OpenStdout:    e.OpenStdout,
		OpenStderr:    e.OpenStderr,
		CanRemove:     e.CanRemove,
		ContainerID:   e.ContainerID,
		DetachKeys:    e.DetachKeys,
		Pid:           e.Pid,
	}, nil
}

// VolumeInspect looks up a volume by name. An error is returned if
// the volume cannot be found.
func (daemon *Daemon) VolumeInspect(name string) (*types.Volume, error) {
	v, err := daemon.volumes.Get(name)
	if err != nil {
		return nil, err
	}
	apiV := volumeToAPIType(v)
	apiV.Mountpoint = v.Path()
	apiV.Status = v.Status()
	return apiV, nil
}

// getDefaultNetworkSettings creates the deprecated structure that holds the information
// about the bridge network for a container.
func (daemon *Daemon) getDefaultNetworkSettings(networks map[string]*network.EndpointSettings) types.DefaultNetworkSettings {
	var settings types.DefaultNetworkSettings

	if defaultNetwork, ok := networks["bridge"]; ok && defaultNetwork.EndpointSettings != nil {
		settings.EndpointID = defaultNetwork.EndpointID
		settings.Gateway = defaultNetwork.Gateway
		settings.GlobalIPv6Address = defaultNetwork.GlobalIPv6Address
		settings.GlobalIPv6PrefixLen = defaultNetwork.GlobalIPv6PrefixLen
		settings.IPAddress = defaultNetwork.IPAddress
		settings.IPPrefixLen = defaultNetwork.IPPrefixLen
		settings.IPv6Gateway = defaultNetwork.IPv6Gateway
		settings.MacAddress = defaultNetwork.MacAddress
	}
	return settings
}
