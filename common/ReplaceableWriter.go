package common

import (
	"io"
	"sync"
)

type ReplaceableWriter struct { // Thread-safe swapout writer.
	mu *sync.Mutex
	w  io.Writer
}

func (r *ReplaceableWriter) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.w.Write(p)
}

func (r *ReplaceableWriter) ReplaceWriter(w io.Writer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.w = w
}
