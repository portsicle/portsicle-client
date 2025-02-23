package client

import (
	"bytes"
	"io"
	"log"

	"github.com/pierrec/lz4"
)

// lz4compress returns the Lempel-Ziv-4 compressed in bytes.
func lz4compress(in []byte) []byte {
	r := bytes.NewReader(in)
	w := &bytes.Buffer{}
	zw := lz4.NewWriter(w) // Creates an LZ4 writer that writes to our buffer

	_, err := io.Copy(zw, r) // Copies data through the compressor
	if err != nil {
		log.Print("Error compressing response body")
		return nil
	}

	if err := zw.Close(); err != nil {
		return nil
	}

	return w.Bytes()
}
