package testlib

import (
	"testing"

	"github.com/goreleaser/goreleaser/pkg/errors"
)

func TestAssertSkipped(t *testing.T) {
	AssertSkipped(t, errors.Skip("test", "skip"))
}
