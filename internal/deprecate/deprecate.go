// Package deprecate provides simple functions to standardize the output
// of deprecation notices on goreleaser
package deprecate

import (
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

const baseURL = "https://goreleaser.com/deprecations#"

// Notice warns the user about the deprecation of the given property
func Notice(property string) {
	// replaces . and _ with -
	url := baseURL + strings.NewReplacer(
		".", "-",
		"_", "-",
	).Replace(property)
	log.Warn(color.New(color.Bold, color.FgHiYellow).Sprintf(
		"DEPRECATED: `%s` should not be used anymore, check %s for more info.",
		property,
		url,
	))
}
