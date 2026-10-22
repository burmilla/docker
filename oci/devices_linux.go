package oci

import (
	"github.com/opencontainers/runc/libcontainer/configs"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Device transforms a libcontainer configs.Device to a specs.LinuxDevice object.
func Device(d *configs.Device) specs.LinuxDevice {
	return specs.LinuxDevice{
		Type:     string(d.Type),
		Path:     d.Path,
		Major:    d.Major,
		Minor:    d.Minor,
		FileMode: fmPtr(int64(d.FileMode)),
		UID:      u32Ptr(int64(d.Uid)),
		GID:      u32Ptr(int64(d.Gid)),
	}
}
