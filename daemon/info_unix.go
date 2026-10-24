//go:build !windows
// +build !windows

package daemon

import (
	"context"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/pkg/sysinfo"
)

// FillPlatformInfo fills the platform related info.
func (daemon *Daemon) FillPlatformInfo(v *types.Info, sysInfo *sysinfo.SysInfo) {
	v.ContainerdCommit.Expected = dockerversion.ContainerdCommitID
	if sv, err := daemon.containerd.GetServerVersion(context.Background()); err == nil {
		v.ContainerdCommit.ID = sv.Revision
	} else {
		logrus.Warnf("failed to retrieve containerd version: %v", err)
		v.ContainerdCommit.ID = "N/A"
	}

	v.RuncCommit.Expected = dockerversion.RuncCommitID
	defaultRuntimeBinary := DefaultRuntimeBinary
	if rv, err := exec.Command(defaultRuntimeBinary, "--version").Output(); err == nil {
		// crun prints "commit: <sha>" on its own line, as runc does.
		for _, line := range strings.Split(string(rv), "\n") {
			if strings.HasPrefix(line, "commit:") {
				v.RuncCommit.ID = strings.TrimSpace(strings.TrimPrefix(line, "commit:"))
				break
			}
		}

		if v.RuncCommit.ID == "" {
			logrus.Warnf("failed to retrieve %s version: unknown output format: %s", defaultRuntimeBinary, string(rv))
			v.RuncCommit.ID = "N/A"
		}
	} else {
		logrus.Warnf("failed to retrieve %s version: %v", defaultRuntimeBinary, err)
		v.RuncCommit.ID = "N/A"
	}
}
