package chrootarchive

import (
	"io"

	"github.com/docker/docker/pkg/archive"
)

// ApplyUncompressedLayer parses a diff in the standard layer format from
// `layer`, and applies it to the directory `dest`. The stream `layer`
// can only be uncompressed.
// Returns the size in bytes of the contents of the layer.
func ApplyUncompressedLayer(dest string, layer io.Reader, options *archive.TarOptions) (int64, error) {
	return applyLayerHandler(dest, layer, options, false)
}
