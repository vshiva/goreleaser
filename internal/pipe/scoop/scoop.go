// Package scoop provides a Pipe that generates a scoop.sh App Manifest and pushes it to a bucket
package scoop

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/goreleaser/goreleaser/internal/artifact"
	"github.com/goreleaser/goreleaser/internal/client"
	"github.com/goreleaser/goreleaser/internal/tmpl"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
)

// XXX

// ErrNoWindows when there is no build for windows (goos doesn't contain windows)
var ErrNoWindows = fmt.Errorf("scoop requires a windows build")

// Pipe for build
type Pipe struct{}

func (Pipe) String() string {
	return "scoop manifest"
}

// Publish scoop manifest
func (Pipe) Publish(ctx *context.Context) error {
	var op errors.Op = "scoop.Publish"
	client, err := client.NewGitHub(ctx)
	if err != nil {
		return err
	}
	return errors.E(op, doRun(ctx, client))
}

// Default sets the pipe defaults
func (Pipe) Default(ctx *context.Context) error {
	if ctx.Config.Scoop.Name == "" {
		ctx.Config.Scoop.Name = ctx.Config.ProjectName
	}
	if ctx.Config.Scoop.CommitAuthor.Name == "" {
		ctx.Config.Scoop.CommitAuthor.Name = "goreleaserbot"
	}
	if ctx.Config.Scoop.CommitAuthor.Email == "" {
		ctx.Config.Scoop.CommitAuthor.Email = "goreleaser@carlosbecker.com"
	}
	if ctx.Config.Scoop.URLTemplate == "" {
		ctx.Config.Scoop.URLTemplate = fmt.Sprintf(
			"%s/%s/%s/releases/download/{{ .Tag }}/{{ .ArtifactName }}",
			ctx.Config.GitHubURLs.Download,
			ctx.Config.Release.GitHub.Owner,
			ctx.Config.Release.GitHub.Name,
		)
	}
	return nil
}

func doRun(ctx *context.Context, client client.Client) error {
	var op errors.Op = "scoop.doRun"
	if ctx.Config.Scoop.Bucket.Name == "" {
		return errors.Skip(op, "scoop section is not configured")
	}
	if ctx.Config.Archive.Format == "binary" {
		return errors.Skip(op, "archive format is binary")
	}

	var archives = ctx.Artifacts.Filter(
		artifact.And(
			artifact.ByGoos("windows"),
			artifact.ByType(artifact.UploadableArchive),
		),
	).List()
	if len(archives) == 0 {
		return errors.Skip(op, ErrNoWindows)
	}

	var path = ctx.Config.Scoop.Name + ".json"

	content, err := buildManifest(ctx, archives)
	if err != nil {
		return errors.Skip(op, err)
	}

	if ctx.SkipPublish {
		return errors.Skip(op, "skip publish enabled")
	}
	if ctx.Config.Release.Draft {
		return errors.Skip(op, "release is marked as draft")
	}
	return client.CreateFile(
		ctx,
		ctx.Config.Scoop.CommitAuthor,
		ctx.Config.Scoop.Bucket,
		content,
		path,
		fmt.Sprintf("Scoop update for %s version %s", ctx.Config.ProjectName, ctx.Git.CurrentTag),
	)
}

// Manifest represents a scoop.sh App Manifest, more info:
// https://github.com/lukesampson/scoop/wiki/App-Manifests
type Manifest struct {
	Version      string              `json:"version"`               // The version of the app that this manifest installs.
	Architecture map[string]Resource `json:"architecture"`          // `architecture`: If the app has 32- and 64-bit versions, architecture can be used to wrap the differences.
	Homepage     string              `json:"homepage,omitempty"`    // `homepage`: The home page for the program.
	License      string              `json:"license,omitempty"`     // `license`: The software license for the program. For well-known licenses, this will be a string like "MIT" or "GPL2". For custom licenses, this should be the URL of the license.
	Description  string              `json:"description,omitempty"` // Description of the app
	Persist      []string            `json:"persist,omitempty"`     // Persist data between updates
}

// Resource represents a combination of a url and a binary name for an architecture
type Resource struct {
	URL  string `json:"url"`  // URL to the archive
	Bin  string `json:"bin"`  // name of binary inside the archive
	Hash string `json:"hash"` // the archive checksum
}

func buildManifest(ctx *context.Context, artifacts []artifact.Artifact) (bytes.Buffer, error) {
	var op errors.Op = "scoop.buildManifest"
	var result bytes.Buffer
	var manifest = Manifest{
		Version:      ctx.Version,
		Architecture: map[string]Resource{},
		Homepage:     ctx.Config.Scoop.Homepage,
		License:      ctx.Config.Scoop.License,
		Description:  ctx.Config.Scoop.Description,
		Persist:      ctx.Config.Scoop.Persist,
	}

	for _, artifact := range artifacts {
		var arch = "64bit"
		if artifact.Goarch == "386" {
			arch = "32bit"
		}

		url, err := tmpl.New(ctx).
			WithArtifact(artifact, map[string]string{}).
			Apply(ctx.Config.Scoop.URLTemplate)
		if err != nil {
			return result, errors.E(op, err)
		}

		sum, err := artifact.Checksum()
		if err != nil {
			return result, errors.E(op, err)
		}

		manifest.Architecture[arch] = Resource{
			URL:  url,
			Bin:  ctx.Config.Builds[0].Binary + ".exe", // TODO: this is wrong
			Hash: sum,
		}
	}

	data, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		return result, errors.E(op, err)
	}
	_, err = result.Write(data)
	return result, errors.E(op, err)
}
