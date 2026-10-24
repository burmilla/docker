//go:build linux
// +build linux

package mount

// MakePrivate ensures a mounted filesystem has the PRIVATE mount option enabled.
// See the supported options in flags.go for further reference.
func MakePrivate(mountPoint string) error {
	return ensureMountedAs(mountPoint, "private")
}

// MakeRPrivate ensures a mounted filesystem has the RPRIVATE mount option
// enabled. See the supported options in flags.go for further reference.
func MakeRPrivate(mountPoint string) error {
	return ensureMountedAs(mountPoint, "rprivate")
}

func ensureMountedAs(mountPoint, options string) error {
	mounted, err := Mounted(mountPoint)
	if err != nil {
		return err
	}

	if !mounted {
		if err := Mount(mountPoint, mountPoint, "none", "bind,rw"); err != nil {
			return err
		}
	}
	if _, err = Mounted(mountPoint); err != nil {
		return err
	}

	return ForceMount("", mountPoint, "none", options)
}
