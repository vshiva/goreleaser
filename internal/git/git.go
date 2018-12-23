// Package git provides an integration with the git command
package git

import (
	"os/exec"
	"strings"

	"github.com/goreleaser/goreleaser/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// IsRepo returns true if current folder is a git repository
func IsRepo() bool {
	out, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Run runs a git command and returns its output or errors
func Run(args ...string) (string, error) {
	var op errors.Op = "git.Run"
	// TODO: use exex.CommandContext here and refactor.
	/* #nosec */
	var cmd = exec.Command("git", args...)
	log.WithField("args", args).Debug("running git")
	bts, err := cmd.CombinedOutput()
	log.WithField("output", string(bts)).
		Debug("git result")
	if err != nil {
		return "", errors.E(op, string(bts))
	}
	return string(bts), nil
}

// Clean the output
func Clean(output string, err error) (string, error) {
	var op errors.Op = "git.Clean"
	output = strings.Replace(strings.Split(output, "\n")[0], "'", "", -1)
	if err != nil {
		err = errors.E(op, strings.TrimSuffix(err.Error(), "\n"))
	}
	return output, nil
}
