package ioutils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"

	"golang.org/x/net/context"
)

type readCloserWrapper struct {
	io.Reader
	closer func() error
}

func (r *readCloserWrapper) Close() error {
	return r.closer()
}

// NewReadCloserWrapper returns a new io.ReadCloser.
func NewReadCloserWrapper(r io.Reader, closer func() error) io.ReadCloser {
	return &readCloserWrapper{
		Reader: r,
		closer: closer,
	}
}

// HashData returns the sha256 sum of src.
func HashData(src io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, src); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// cancelReadCloser wraps an io.ReadCloser with a context for cancelling read
// operations.
type cancelReadCloser struct {
	cancel func()
	pR     *io.PipeReader // Stream to read from
	pW     *io.PipeWriter
}

// NewCancelReadCloser creates a wrapper that closes the ReadCloser when the
// context is cancelled. The returned io.ReadCloser must be closed when it is
// no longer needed.
func NewCancelReadCloser(ctx context.Context, in io.ReadCloser) io.ReadCloser {
	pR, pW := io.Pipe()

	// Create a context used to signal when the pipe is closed
	doneCtx, cancel := context.WithCancel(context.Background())

	p := &cancelReadCloser{
		cancel: cancel,
		pR:     pR,
		pW:     pW,
	}

	go func() {
		_, err := io.Copy(pW, in)
		select {
		case <-ctx.Done():
			// If the context was closed, p.closeWithError
			// was already called. Calling it again would
			// change the error that Read returns.
		default:
			p.closeWithError(err)
		}
		in.Close()
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				p.closeWithError(ctx.Err())
			case <-doneCtx.Done():
				return
			}
		}
	}()

	return p
}

// Read wraps the Read method of the pipe that provides data from the wrapped
// ReadCloser.
func (p *cancelReadCloser) Read(buf []byte) (n int, err error) {
	return p.pR.Read(buf)
}

// closeWithError closes the wrapper and its underlying reader. It will
// cause future calls to Read to return err.
func (p *cancelReadCloser) closeWithError(err error) {
	p.pW.CloseWithError(err)
	p.cancel()
}

// Close closes the wrapper its underlying reader. It will cause
// future calls to Read to return io.EOF.
func (p *cancelReadCloser) Close() error {
	p.closeWithError(io.EOF)
	return nil
}
