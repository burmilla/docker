package sysinfo

import "github.com/docker/docker/pkg/parsers"

// SysInfo stores information about which features a kernel supports.
// TODO Windows: Factor out platform specific capabilities.
type SysInfo struct {
	// Whether the kernel supports AppArmor or not
	AppArmor bool
	// Whether the kernel supports Seccomp or not
	Seccomp bool

	// Whether IPv4 forwarding is supported or not, if this was disabled, networking will not work
	IPv4ForwardingDisabled bool

	// Whether bridge-nf-call-iptables is supported or not
	BridgeNFCallIPTablesDisabled bool

	// Whether bridge-nf-call-ip6tables is supported or not
	BridgeNFCallIP6TablesDisabled bool
}

func isCpusetListAvailable(provided, available string) (bool, error) {
	parsedProvided, err := parsers.ParseUintList(provided)
	if err != nil {
		return false, err
	}
	parsedAvailable, err := parsers.ParseUintList(available)
	if err != nil {
		return false, err
	}
	for k := range parsedProvided {
		if !parsedAvailable[k] {
			return false, nil
		}
	}
	return true, nil
}

// Returns bit count of 1, used by NumCPU
func popcnt(x uint64) (n byte) {
	x -= (x >> 1) & 0x5555555555555555
	x = (x>>2)&0x3333333333333333 + x&0x3333333333333333
	x += x >> 4
	x &= 0x0f0f0f0f0f0f0f0f
	x *= 0x0101010101010101
	return byte(x >> 56)
}
