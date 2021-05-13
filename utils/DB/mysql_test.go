package DB

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestBuildDSN(t *testing.T) {
	spew.Dump(buildDSN())
}
