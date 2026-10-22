package daemon

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/stringid"
	"github.com/pkg/errors"
)

// loadOrCreateID reads the daemon ID from path, generating and persisting a
// random one on first start. It replaces the libtrust key ID of older versions.
func loadOrCreateID(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err == nil {
		if id := strings.TrimSpace(string(b)); id != "" {
			return id, nil
		}
	} else if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "error reading daemon ID file %s", path)
	}
	id := stringid.GenerateRandomID()
	if err := ioutils.AtomicWriteFile(path, []byte(id+"\n"), 0600); err != nil {
		return "", errors.Wrapf(err, "error writing daemon ID file %s", path)
	}
	return id, nil
}
