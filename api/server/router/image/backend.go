package image

import (
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"golang.org/x/net/context"
)

// Backend is all the methods that need to be implemented
// to provide image specific functionality.
type Backend interface {
	imageBackend
	importExportBackend
	registryBackend
}

type imageBackend interface {
	ImageDelete(imageRef string, force, prune bool) ([]types.ImageDeleteResponseItem, error)
	ImageHistory(imageName string) ([]*image.HistoryResponseItem, error)
	Images(imageFilters filters.Args, all bool, withExtraAttrs bool) ([]*types.ImageSummary, error)
	LookupImage(name string) (*types.ImageInspect, error)
	TagImage(imageName, repository, tag string) error
	ImagesPrune(ctx context.Context, pruneFilters filters.Args) (*types.ImagesPruneReport, error)
}

type importExportBackend interface {
	LoadImage(inTar io.ReadCloser, outStream io.Writer, quiet bool) error
	ExportImage(names []string, outStream io.Writer) error
}

type registryBackend interface {
	PullImage(ctx context.Context, image, tag, platform string, metaHeaders map[string][]string, authConfig *types.AuthConfig, outStream io.Writer) error
}
