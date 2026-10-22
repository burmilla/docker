package types

import (
	"bufio"
	"io"
	"net"

	"github.com/docker/docker/api/types/filters"
)

// ContainerAttachOptions holds parameters to attach to a container.
type ContainerAttachOptions struct {
	Stream     bool
	Stdin      bool
	Stdout     bool
	Stderr     bool
	DetachKeys string
	Logs       bool
}

// ContainerExecInspect holds information returned by exec inspect.
type ContainerExecInspect struct {
	ExecID      string
	ContainerID string
	Running     bool
	ExitCode    int
	Pid         int
}

// ContainerListOptions holds parameters to list containers with.
type ContainerListOptions struct {
	Quiet   bool
	Size    bool
	All     bool
	Latest  bool
	Since   string
	Before  string
	Limit   int
	Filters filters.Args
}

// ContainerLogsOptions holds parameters to filter logs with.
type ContainerLogsOptions struct {
	ShowStdout bool
	ShowStderr bool
	Since      string
	Timestamps bool
	Follow     bool
	Tail       string
	Details    bool
}

// ContainerRemoveOptions holds parameters to remove containers.
type ContainerRemoveOptions struct {
	RemoveVolumes bool
	RemoveLinks   bool
	Force         bool
}

// ContainerStartOptions holds parameters to start containers.
type ContainerStartOptions struct {
	CheckpointID  string
	CheckpointDir string
}

// CopyToContainerOptions holds information
// about files to copy into a container
type CopyToContainerOptions struct {
	AllowOverwriteDirWithFile bool
	CopyUIDGID                bool
}

// EventsOptions holds parameters to filter events with.
type EventsOptions struct {
	Since   string
	Until   string
	Filters filters.Args
}

// NetworkListOptions holds parameters to filter the list of networks with.
type NetworkListOptions struct {
	Filters filters.Args
}

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Conn   net.Conn
	Reader *bufio.Reader
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() {
	h.Conn.Close()
}

// CloseWriter is an interface that implements structs
// that close input streams to prevent from writing.
type CloseWriter interface {
	CloseWrite() error
}

// CloseWrite closes a readWriter for writing.
func (h *HijackedResponse) CloseWrite() error {
	if conn, ok := h.Conn.(CloseWriter); ok {
		return conn.CloseWrite()
	}
	return nil
}

// ImageCreateOptions holds information to create images.
type ImageCreateOptions struct {
	RegistryAuth string // RegistryAuth is the base64 encoded credentials for the registry
}

// ImageListOptions holds parameters to filter the list of images with.
type ImageListOptions struct {
	All     bool
	Filters filters.Args
}

// ImageLoadResponse returns information to the client about a load process.
type ImageLoadResponse struct {
	// Body must be closed to avoid a resource leak
	Body io.ReadCloser
	JSON bool
}

// ImagePullOptions holds information to pull images.
type ImagePullOptions struct {
	All           bool
	RegistryAuth  string // RegistryAuth is the base64 encoded credentials for the registry
	PrivilegeFunc RequestPrivilegeFunc
}

// RequestPrivilegeFunc is a function interface that
// clients can supply to retry operations after
// getting an authorization error.
// This function returns the registry authentication
// header value in base 64 format, or an error
// if the privilege request fails.
type RequestPrivilegeFunc func() (string, error)

// ImageRemoveOptions holds parameters to remove images.
type ImageRemoveOptions struct {
	Force         bool
	PruneChildren bool
}

// ResizeOptions holds parameters to resize a tty.
// It can be used to resize container ttys and
// exec process ttys too.
type ResizeOptions struct {
	Height uint
	Width  uint
}
