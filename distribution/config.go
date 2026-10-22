package distribution

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/distribution/metadata"
	"github.com/docker/docker/distribution/xfer"
	"github.com/docker/docker/image"
	"github.com/docker/docker/layer"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/pkg/system"
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
	// RequireSchema2 ensures that only schema2 manifests are used.
	RequireSchema2 bool
}

// ImagePullConfig stores pull configuration.
type ImagePullConfig struct {
	Config

	// DownloadManager manages concurrent pulls.
	DownloadManager RootFSDownloadManager
	// Schema2Types is the valid schema2 configuration types allowed
	// by the pull operation.
	Schema2Types []string
	// Platform is the requested platform of the image being pulled to ensure it can be validated
	// when the host platform supports multiple image operating systems.
	Platform string
}

// ImageConfigStore handles storing and getting image configurations
// by digest. Allows getting an image configurations rootfs from the
// configuration.
type ImageConfigStore interface {
	Put([]byte) (digest.Digest, error)
	Get(digest.Digest) ([]byte, error)
	RootFSAndPlatformFromConfig([]byte) (*image.RootFS, layer.Platform, error)
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

func (s *imageConfigStore) RootFSAndPlatformFromConfig(c []byte) (*image.RootFS, layer.Platform, error) {
	var unmarshalledConfig image.Image
	if err := json.Unmarshal(c, &unmarshalledConfig); err != nil {
		return nil, "", err
	}

	// fail immediately on Windows when downloading a non-Windows image
	// and vice versa. Exception on Windows if Linux Containers are enabled.
	if runtime.GOOS == "windows" && unmarshalledConfig.OS == "linux" && !system.LCOWSupported() {
		return nil, "", fmt.Errorf("image operating system %q cannot be used on this platform", unmarshalledConfig.OS)
	} else if runtime.GOOS != "windows" && unmarshalledConfig.OS == "windows" {
		return nil, "", fmt.Errorf("image operating system %q cannot be used on this platform", unmarshalledConfig.OS)
	}

	platform := ""
	if runtime.GOOS == "windows" {
		platform = unmarshalledConfig.OS
	}
	return unmarshalledConfig.RootFS, layer.Platform(platform), nil
}
