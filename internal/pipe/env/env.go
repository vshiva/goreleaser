// Package env implements the Pipe interface providing validation of
// missing environment variables needed by the release process.
package env

import (
	"bufio"
	"fmt"
	"os"

	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/goreleaser/goreleaser/pkg/errors"
	homedir "github.com/mitchellh/go-homedir"
)

// ErrMissingToken indicates an error when GITHUB_TOKEN is missing in the environment
var ErrMissingToken = fmt.Errorf("missing GITHUB_TOKEN")

// Pipe for env
type Pipe struct{}

func (Pipe) String() string {
	return "loading environment variables"
}

// Default sets the pipe defaults
func (Pipe) Default(ctx *context.Context) error {
	var env = &ctx.Config.EnvFiles
	if env.GitHubToken == "" {
		env.GitHubToken = "~/.config/goreleaser/github_token"
	}
	return nil
}

// Run the pipe
func (Pipe) Run(ctx *context.Context) error {
	const op errors.Op = "env.Run"
	token, err := loadEnv("GITHUB_TOKEN", ctx.Config.EnvFiles.GitHubToken)
	ctx.Token = token
	if ctx.SkipPublish {
		return errors.Skip(op, "publishing is disabled")
	}
	if ctx.Config.Release.Disable {
		return errors.Skip(op, "release pipe is disabled")
	}
	if ctx.Token == "" && err == nil {
		return errors.E(op, ErrMissingToken)
	}
	return errors.E(op, err)
}

func loadEnv(env, path string) (string, error) {
	val := os.Getenv(env)
	if val != "" {
		return val, nil
	}
	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path) // #nosec
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	bts, _, err := bufio.NewReader(f).ReadLine()
	return string(bts), err
}
