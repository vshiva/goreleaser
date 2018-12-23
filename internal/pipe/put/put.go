// Package put provides a Pipe that push using HTTP PUT
package put

import (
	"fmt"
	h "net/http"

	"github.com/goreleaser/goreleaser/internal/http"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
)

// Pipe for http publishing
type Pipe struct{}

// String returns the description of the pipe
func (Pipe) String() string {
	return "HTTP PUT"
}

// Default sets the pipe defaults
func (Pipe) Default(ctx *context.Context) error {
	return http.Defaults(ctx.Config.Puts)
}

// Publish artifacts
func (Pipe) Publish(ctx *context.Context) error {
	var op errors.Op = "http.Publish"
	if len(ctx.Config.Puts) == 0 {
		return errors.Skip(op, "put section is not configured")
	}

	// Check requirements for every instance we have configured.
	// If not fulfilled, we can skip this pipeline
	for _, instance := range ctx.Config.Puts {
		instance := instance
		if err := http.CheckConfig(ctx, &instance, "put"); err != nil {
			return errors.E(op, err)
		}
	}

	return http.Upload(ctx, ctx.Config.Puts, "put", func(res *h.Response) error {
		if c := res.StatusCode; c < 200 || 299 < c {
			return errors.E(op, fmt.Sprintf("unexpected http response status: %s", res.Status))
		}
		return nil
	})

}
