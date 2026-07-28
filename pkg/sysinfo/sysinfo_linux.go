package sysinfo

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

const (
	// SeccompModeFilter refers to the syscall argument SECCOMP_MODE_FILTER.
	SeccompModeFilter = uintptr(2)
)

// New returns a new SysInfo, using the filesystem to detect which features
// the kernel supports. If `quiet` is `false` warnings are printed in logs
// whenever an error occurs or misconfigurations are present.
func New(quiet bool) *SysInfo {
	sysInfo := &SysInfo{}

	sysInfo.IPv4ForwardingDisabled = !readProcBool("/proc/sys/net/ipv4/ip_forward")
	sysInfo.BridgeNFCallIPTablesDisabled = !readProcBool("/proc/sys/net/bridge/bridge-nf-call-iptables")
	sysInfo.BridgeNFCallIP6TablesDisabled = !readProcBool("/proc/sys/net/bridge/bridge-nf-call-ip6tables")

	// Check if AppArmor is supported.
	if _, err := os.Stat("/sys/kernel/security/apparmor"); !os.IsNotExist(err) {
		sysInfo.AppArmor = true
	}

	// Check if Seccomp is supported, via CONFIG_SECCOMP.
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_GET_SECCOMP, 0, 0); err != syscall.EINVAL {
		// Make sure the kernel has CONFIG_SECCOMP_FILTER.
		if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_SECCOMP, SeccompModeFilter, 0); err != syscall.EINVAL {
			sysInfo.Seccomp = true
		}
	}

	return sysInfo
}

func readProcBool(path string) bool {
	val, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(val)) == "1"
}
