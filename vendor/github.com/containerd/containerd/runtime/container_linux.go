package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/containerd/containerd/specs"
	ocs "github.com/opencontainers/runtime-spec/specs-go"
)

func (c *container) Pids() ([]int, error) {
	var pids []int
	args := c.runtimeArgs
	args = append(args, "ps", "--format=json", c.id)
	out, err := exec.Command(c.runtime, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %q", err.Error(), out)
	}
	if err := json.Unmarshal(out, &pids); err != nil {
		return nil, err
	}
	return pids, nil
}

func u64Ptr(i uint64) *uint64 { return &i }
func i64Ptr(i int64) *int64   { return &i }

func (c *container) UpdateResources(r *Resource) error {
	sr := ocs.LinuxResources{
		Memory: &ocs.LinuxMemory{
			Limit:       u64Ptr(uint64(r.Memory)),
			Reservation: u64Ptr(uint64(r.MemoryReservation)),
			Swap:        u64Ptr(uint64(r.MemorySwap)),
			Kernel:      u64Ptr(uint64(r.KernelMemory)),
			KernelTCP:   u64Ptr(uint64(r.KernelTCPMemory)),
		},
		CPU: &ocs.LinuxCPU{
			Shares: u64Ptr(uint64(r.CPUShares)),
			Quota:  i64Ptr(int64(r.CPUQuota)),
			Period: u64Ptr(uint64(r.CPUPeriod)),
			Cpus:   r.CpusetCpus,
			Mems:   r.CpusetMems,
		},
		BlockIO: &ocs.LinuxBlockIO{
			Weight: &r.BlkioWeight,
		},
		Pids: &ocs.LinuxPids{
			Limit: r.PidsLimit,
		},
	}

	srStr := bytes.NewBuffer(nil)
	if err := json.NewEncoder(srStr).Encode(&sr); err != nil {
		return err
	}

	args := c.runtimeArgs
	args = append(args, "update", "-r", "-", c.id)
	cmd := exec.Command(c.runtime, args...)
	cmd.Stdin = srStr
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(b))
	}
	return nil
}

func getRootIDs(s *specs.Spec) (int, int, error) {
	if s == nil {
		return 0, 0, nil
	}
	var hasUserns bool
	for _, ns := range s.Linux.Namespaces {
		if ns.Type == ocs.UserNamespace {
			hasUserns = true
			break
		}
	}
	if !hasUserns {
		return 0, 0, nil
	}
	uid := hostIDFromMap(0, s.Linux.UIDMappings)
	gid := hostIDFromMap(0, s.Linux.GIDMappings)
	return uid, gid, nil
}
