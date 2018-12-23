// Package snapshot provides the snapshoting functionality to goreleaser.
package snapshot

import (
	"fmt"

	"github.com/goreleaser/goreleaser/internal/tmpl"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
)

// Pipe for checksums
type Pipe struct{}

func (Pipe) String() string {
	return "snapshoting"
}

// Default sets the pipe defaults
func (Pipe) Default(ctx *context.Context) error {
	if ctx.Config.Snapshot.NameTemplate == "" {
		ctx.Config.Snapshot.NameTemplate = "SNAPSHOT-{{ .ShortCommit }}"
	}
	return nil
}

func (Pipe) Run(ctx *context.Context) error {
	var op errors.Op = "snapshot.Run"
	if !ctx.Snapshot {
		return errors.Skip(op, "not a snapshot")
	}
	name, err := tmpl.New(ctx).Apply(ctx.Config.Snapshot.NameTemplate)
	if err != nil {
		return errors.E(op, err, "failed to generate snapshot name")
	}
	if name == "" {
		return errors.E(op, fmt.Errorf("empty snapshot name"))
	}
	ctx.Version = name
	return nil
}
