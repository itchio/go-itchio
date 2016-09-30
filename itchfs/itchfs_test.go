package itchfs_test

import (
	"testing"

	"github.com/itchio/go-itchio/itchfs"
	"github.com/itchio/wharf/eos"
)

func Test_Register(t *testing.T) {
	eos.RegisterHandler(&itchfs.ItchFS{})
}
