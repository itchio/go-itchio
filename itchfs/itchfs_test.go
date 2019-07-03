package itchfs

import (
	"net/http"
	"testing"

	"github.com/itchio/httpkit/eos"
	"github.com/stretchr/testify/assert"
)

func Test_Register(t *testing.T) {
	ifs := &ItchFS{}
	assert.NoError(t, eos.RegisterHandler(ifs))
	defer eos.DeregisterHandler(ifs)
	assert.Error(t, eos.RegisterHandler(ifs))
}

func Test_Renewal(t *testing.T) {
	res := &http.Response{
		StatusCode: 400,
	}
	assert.True(t, needsRenewal(res, nil))

	res.StatusCode = 200
	assert.False(t, needsRenewal(res, nil))
}
