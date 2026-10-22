package daemon

import (
	"errors"
	"runtime"

	"github.com/docker/docker/pkg/archive"
)

// ContainerChanges returns a list of container fs changes
func (daemon *Daemon) ContainerChanges(name string) ([]archive.Change, error) {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "windows" && container.IsRunning() {
		return nil, errors.New("Windows does not support diff of a running container")
	}

	container.Lock()
	defer container.Unlock()
	c, err := container.RWLayer.Changes()
	if err != nil {
		return nil, err
	}
	return c, nil
}
