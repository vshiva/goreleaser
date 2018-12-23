// Package publish contains the publishing pipe.
package publish

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/goreleaser/goreleaser/internal/pipe/artifactory"
	"github.com/goreleaser/goreleaser/internal/pipe/brew"
	"github.com/goreleaser/goreleaser/internal/pipe/docker"
	"github.com/goreleaser/goreleaser/internal/pipe/put"
	"github.com/goreleaser/goreleaser/internal/pipe/release"
	"github.com/goreleaser/goreleaser/internal/pipe/s3"
	"github.com/goreleaser/goreleaser/internal/pipe/scoop"
	"github.com/goreleaser/goreleaser/internal/pipe/snapcraft"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
)

// Pipe that publishes artifacts
type Pipe struct{}

func (Pipe) String() string {
	return "publishing"
}

// Publisher should be implemented by pipes that want to publish artifacts
type Publisher interface {
	fmt.Stringer

	// Default sets the configuration defaults
	Publish(ctx *context.Context) error
}

// nolint: gochecknoglobals
var publishers = []Publisher{
	s3.Pipe{},
	put.Pipe{},
	artifactory.Pipe{},
	docker.Pipe{},
	snapcraft.Pipe{},
	// This should be one of the last steps
	release.Pipe{},
	// brew and scoop use the release URL, so, they should be last
	brew.Pipe{},
	scoop.Pipe{},
}

// Run the pipe
func (Pipe) Run(ctx *context.Context) error {
	var op errors.Op = "publish.Run"
	if ctx.SkipPublish {
		return errors.Skip(op, "skip publish")
	}
	for _, publisher := range publishers {
		log.Infof(color.New(color.Bold).Sprint(publisher.String()))
		if err := handle(publisher.Publish(ctx)); err != nil {
			return errors.E(op, err, fmt.Sprintf("%s: failed to publish artifacts", publisher.String()))
		}
	}
	return nil
}

// XXX
// TODO: for now this is duplicated, we should have better error handling
// eventually.
func handle(err error) error {
	if err == nil {
		return nil
	}
	if errors.IsSkip(err) {
		log.WithField("reason", err.Error()).WithField("ops", errors.Ops(err)).Warn("skipped")
		return nil
	}
	return err
}
