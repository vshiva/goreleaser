package testlib

import (
	"testing"

	"github.com/goreleaser/goreleaser/pkg/errors"
)

func TestAssertSkipped(t *testing.T) {
	var op errors.Op = "test"
	AssertSkipped(t, errors.Skip("test", "skip"))
	AssertSkipped(t, errors.E(op, errors.Skip(op, "skip")))
}
