package main

import (
	"bytes"
	"io"
)

// readAll reads a body fully, returning an error if the underlying
// MaxBytesReader tripped. io.ReadAll's own error is what carries that, so this
// exists to keep the intent obvious at the call site rather than to add logic.
func readAll(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
