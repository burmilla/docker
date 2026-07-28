//go:build !windows
// +build !windows

package main

import (
	"github.com/docker/docker/integration-cli/request"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestAPIUpdateContainer(c *check.C) {
	testRequires(c, DaemonIsLinux)
	testRequires(c, memoryLimitSupport)
	testRequires(c, swapMemorySupport)

	name := "apiUpdateContainer"
	hostConfig := map[string]interface{}{
		"Memory":     314572800,
		"MemorySwap": 524288000,
	}
	dockerCmd(c, "run", "-d", "--name", name, "-m", "200M", "busybox", "top")
	_, _, err := request.SockRequest("POST", "/containers/"+name+"/update", hostConfig, daemonHost())
	c.Assert(err, check.IsNil)
}
