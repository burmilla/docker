package volumedrivers

import (
	"fmt"
	"sort"
	"sync"

	"github.com/docker/docker/volume"
)

var drivers = &driverExtpoint{
	extensions: make(map[string]volume.Driver),
}

type driverExtpoint struct {
	extensions map[string]volume.Driver
	sync.Mutex
}

// Register associates the given driver to the given name, checking if
// the name is already associated
func Register(extension volume.Driver, name string) bool {
	if name == "" {
		return false
	}

	drivers.Lock()
	defer drivers.Unlock()

	_, exists := drivers.extensions[name]
	if exists {
		return false
	}

	if err := validateDriver(extension); err != nil {
		return false
	}

	drivers.extensions[name] = extension

	return true
}

// Unregister dissociates the name from its driver, if the association exists.
func Unregister(name string) bool {
	drivers.Lock()
	defer drivers.Unlock()

	_, exists := drivers.extensions[name]
	if !exists {
		return false
	}
	delete(drivers.extensions, name)
	return true
}

// lookup returns the driver associated with the given name.
func lookup(name string) (volume.Driver, error) {
	drivers.Lock()
	ext, ok := drivers.extensions[name]
	drivers.Unlock()
	if ok {
		return ext, nil
	}
	return nil, fmt.Errorf("Error looking up volume driver %s", name)
}

func validateDriver(vd volume.Driver) error {
	scope := vd.Scope()
	if scope != volume.LocalScope && scope != volume.GlobalScope {
		return fmt.Errorf("Driver %q provided an invalid capability scope: %s", vd.Name(), scope)
	}
	return nil
}

// GetDriver returns a volume driver by its name.
// If the driver is empty, it looks for the local driver.
func GetDriver(name string) (volume.Driver, error) {
	if name == "" {
		name = volume.DefaultDriverName
	}
	return lookup(name)
}

// CreateDriver returns a volume driver by its name.
// If the driver is empty, it looks for the local driver.
func CreateDriver(name string) (volume.Driver, error) {
	return GetDriver(name)
}

// RemoveDriver returns a volume driver by its name.
// If the driver is empty, it looks for the local driver.
func RemoveDriver(name string) (volume.Driver, error) {
	return GetDriver(name)
}

// GetDriverList returns list of volume drivers registered.
// If no driver is registered, empty string list will be returned.
func GetDriverList() []string {
	var driverList []string
	drivers.Lock()
	for driverName := range drivers.extensions {
		driverList = append(driverList, driverName)
	}
	drivers.Unlock()
	sort.Strings(driverList)
	return driverList
}

// GetAllDrivers lists all the registered drivers
func GetAllDrivers() ([]volume.Driver, error) {
	var ds []volume.Driver

	drivers.Lock()
	defer drivers.Unlock()

	for _, d := range drivers.extensions {
		ds = append(ds, d)
	}
	return ds, nil
}
