package git

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/goreleaser/goreleaser/internal/deprecate"
	"github.com/goreleaser/goreleaser/internal/git"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Pipe that sets up git state
type Pipe struct{}

func (Pipe) String() string {
	return "getting and validating git state"
}

// Run the pipe
func (Pipe) Run(ctx *context.Context) error {
	var op errors.Op = "git.Run"
	if _, err := exec.LookPath("git"); err != nil {
		return errors.E(op, ErrNoGit)
	}
	if ctx.Config.Git.ShortHash {
		deprecate.Notice("git.short_hash")
	}
	info, err := getInfo(ctx)
	if err != nil {
		return errors.E(op, ErrNoGit)
	}
	ctx.Git = info
	log.Infof("releasing %s, commit %s", info.CurrentTag, info.Commit)
	ctx.Version = strings.TrimPrefix(ctx.Git.CurrentTag, "v")
	return errors.E(op, validate(ctx))
}

// nolint: gochecknoglobals
var fakeInfo = context.GitInfo{
	CurrentTag:  "v0.0.0",
	Commit:      "none",
	ShortCommit: "none",
	FullCommit:  "none",
}

func getInfo(ctx *context.Context) (context.GitInfo, error) {
	var op errors.Op = "git.getInfo"
	if !git.IsRepo() && ctx.Snapshot {
		log.Warn("accepting to run without a git repo because this is a snapshot")
		return fakeInfo, nil
	}
	if !git.IsRepo() {
		return context.GitInfo{}, errors.E(op, ErrNotRepository)
	}
	info, err := getGitInfo(ctx)
	if err != nil && ctx.Snapshot {
		log.WithError(err).Warn("ignoring errors because this is a snapshot")
		if info.Commit == "" {
			info = fakeInfo
		}
		return info, nil
	}
	return info, errors.E(op, err)
}

func getGitInfo(ctx *context.Context) (context.GitInfo, error) {
	var op errors.Op = "git.getGitInfo"
	short, err := getShortCommit()
	if err != nil {
		return context.GitInfo{}, errors.E(op, err, "couldn't get current commit")
	}
	full, err := getFullCommit()
	if err != nil {
		return context.GitInfo{}, errors.E(op, err, "couldn't get current commit")
	}
	var commit = full
	if ctx.Config.Git.ShortHash {
		commit = short
	}
	url, err := getURL()
	if err != nil {
		return context.GitInfo{}, errors.E(op, err, "couldn't get remote URL")
	}
	tag, err := getTag()
	if err != nil {
		return context.GitInfo{
			Commit:      commit,
			FullCommit:  full,
			ShortCommit: short,
			URL:         url,
			CurrentTag:  "v0.0.0",
		}, errors.E(op, ErrNoTag)
	}
	return context.GitInfo{
		CurrentTag:  tag,
		Commit:      commit,
		FullCommit:  full,
		ShortCommit: short,
		URL:         url,
	}, nil
}

func validate(ctx *context.Context) error {
	var op errors.Op = "git.validate"
	if ctx.Snapshot {
		return errors.Skip(op, "snapshot is enabled")
	}
	if ctx.SkipValidate {
		return errors.Skip(op, "skip validation is enabled")
	}
	out, err := git.Run("status", "--porcelain")
	if strings.TrimSpace(out) != "" || err != nil {
		return errors.E(op, ErrDirty{status: out})
	}
	if !regexp.MustCompile("^[0-9.]+").MatchString(ctx.Version) {
		return errors.E(op, ErrInvalidVersionFormat{version: ctx.Version})
	}
	_, err = git.Clean(git.Run("describe", "--exact-match", "--tags", "--match", ctx.Git.CurrentTag))
	if err != nil {
		return errors.E(op, ErrWrongRef{
			commit: ctx.Git.Commit,
			tag:    ctx.Git.CurrentTag,
		})
	}
	return nil
}

func getShortCommit() (string, error) {
	return git.Clean(git.Run("show", "--format='%h'", "HEAD"))
}

func getFullCommit() (string, error) {
	return git.Clean(git.Run("show", "--format='%H'", "HEAD"))
}

func getTag() (string, error) {
	return git.Clean(git.Run("describe", "--tags", "--abbrev=0"))
}

func getURL() (string, error) {
	return git.Clean(git.Run("ls-remote", "--get-url"))
}
