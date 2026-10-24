package distribution

import (
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/distribution/metadata"
	"github.com/docker/docker/distribution/xfer"
	"github.com/docker/docker/image"
	"github.com/docker/docker/layer"
	"github.com/docker/docker/pkg/progress"
	refstore "github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/opencontainers/go-digest"
	"golang.org/x/net/context"
)

// Config stores configuration for communicating
// with a registry.
type Config struct {
	// MetaHeaders stores HTTP headers with metadata about the image
	MetaHeaders map[string][]string
	// AuthConfig holds authentication credentials for authenticating with
	// the registry.
	AuthConfig *types.AuthConfig
	// ProgressOutput is the interface for showing the status of the pull
	// operation.
	ProgressOutput progress.Output
	// RegistryService is the registry service to use for TLS configuration
	// and endpoint lookup.
	RegistryService registry.Service
	// ImageEventLogger notifies events for a given image
	ImageEventLogger func(id, name, action string)
	// MetadataStore is the storage backend for distribution-specific
	// metadata.
	MetadataStore metadata.Store
	// ImageStore manages images.
	ImageStore ImageConfigStore
	// ReferenceStore manages tags. This value is optional, when excluded
	// content will not be tagged.
	ReferenceStore refstore.Store
}

// ImagePullConfig stores pull configuration.
type ImagePullConfig struct {
	Config

	// DownloadManager manages concurrent pulls.
	DownloadManager RootFSDownloadManager
	// Schema2Types is the valid schema2 configuration types allowed
	// by the pull operation.
	Schema2Types []string
}

// ImageConfigStore handles storing and getting image configurations
// by digest. Allows getting an image configurations rootfs from the
// configuration.
type ImageConfigStore interface {
	Put([]byte) (digest.Digest, error)
	Get(digest.Digest) ([]byte, error)
	RootFSFromConfig([]byte) (*image.RootFS, error)
}

// RootFSDownloadManager handles downloading of the rootfs
type RootFSDownloadManager interface {
	// Download downloads the layers into the given initial rootfs and
	// returns the final rootfs.
	// Given progress output to track download progress
	// Returns function to release download resources
	Download(ctx context.Context, initialRootFS image.RootFS, platform layer.Platform, layers []xfer.DownloadDescriptor, progressOutput progress.Output) (image.RootFS, func(), error)
}

type imageConfigStore struct {
	image.Store
}

// NewImageConfigStoreFromStore returns an ImageConfigStore backed
// by an image.Store for container images.
func NewImageConfigStoreFromStore(is image.Store) ImageConfigStore {
	return &imageConfigStore{
		Store: is,
	}
}

func (s *imageConfigStore) Put(c []byte) (digest.Digest, error) {
	id, err := s.Store.Create(c)
	return digest.Digest(id), err
}

func (s *imageConfigStore) Get(d digest.Digest) ([]byte, error) {
	img, err := s.Store.Get(image.IDFromDigest(d))
	if err != nil {
		return nil, err
	}
	return img.RawJSON(), nil
}

func (s *imageConfigStore) RootFSFromConfig(c []byte) (*image.RootFS, error) {
	var unmarshalledConfig image.Image
	if err := json.Unmarshal(c, &unmarshalledConfig); err != nil {
		return nil, err
	}

	if unmarshalledConfig.OS == "windows" {
		return nil, fmt.Errorf("image operating system %q cannot be used on this platform", unmarshalledConfig.OS)
	}
	return unmarshalledConfig.RootFS, nil
}
