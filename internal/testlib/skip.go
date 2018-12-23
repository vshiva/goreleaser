package testlib

import (
	"testing"

	"github.com/goreleaser/goreleaser/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// AssertSkipped asserts that a pipe was skipped
func AssertSkipped(t *testing.T, err error) {
	assert.True(t, errors.IsSkip(err), "expected an errors.KindPipeSkipped but got %v", err)
}
