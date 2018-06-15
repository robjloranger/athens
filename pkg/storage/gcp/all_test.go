package gcp

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/api/option"
)

type GcpTests struct {
	suite.Suite
	options option.ClientOption
}

func (g *GcpTests) SetupTest() {
	g.options = option.WithoutAuthentication()
}

func TestGcpStorage(t *testing.T) {
	suite.Run(t, new(GcpTests))
}
