package listeners

import (
	"fmt"
	"net"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go-connections/sockets"
)

// Init creates a new unix socket listener for the server.
func Init(addr, socketGroup string) ([]net.Listener, error) {
	gid, err := lookupGID(socketGroup)
	if err != nil {
		if socketGroup != "" {
			if socketGroup != defaultSocketGroup {
				return nil, err
			}
			logrus.Warnf("could not change group %s to %s: %v", addr, defaultSocketGroup, err)
		}
		gid = os.Getgid()
	}
	l, err := sockets.NewUnixSocket(addr, gid)
	if err != nil {
		return nil, fmt.Errorf("can't create unix socket %s: %v", addr, err)
	}
	return []net.Listener{l}, nil
}
