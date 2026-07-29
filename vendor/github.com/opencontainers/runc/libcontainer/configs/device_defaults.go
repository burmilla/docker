// +build linux freebsd

package configs

var (
	// DefaultSimpleDevices are devices that are to be both allowed and created.
	DefaultSimpleDevices = []*Device{
		// /dev/null and zero
		{
			Path:        "/dev/null",
			Type:        'c',
			Major:       1,
			Minor:       3,
			Permissions: "rwm",
			FileMode:    0666,
		},
		{
			Path:        "/dev/zero",
			Type:        'c',
			Major:       1,
			Minor:       5,
			Permissions: "rwm",
			FileMode:    0666,
		},

		{
			Path:        "/dev/full",
			Type:        'c',
			Major:       1,
			Minor:       7,
			Permissions: "rwm",
			FileMode:    0666,
		},

		// consoles and ttys
		{
			Path:        "/dev/tty",
			Type:        'c',
			Major:       5,
			Minor:       0,
			Permissions: "rwm",
			FileMode:    0666,
		},

		// /dev/urandom,/dev/random
		{
			Path:        "/dev/urandom",
			Type:        'c',
			Major:       1,
			Minor:       9,
			Permissions: "rwm",
			FileMode:    0666,
		},
		{
			Path:        "/dev/random",
			Type:        'c',
			Major:       1,
			Minor:       8,
			Permissions: "rwm",
			FileMode:    0666,
		},
	}
	DefaultAutoCreatedDevices = append([]*Device{}, DefaultSimpleDevices...)
)
