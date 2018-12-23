package errlog

import (
	"strings"

	"github.com/goreleaser/goreleaser/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SystemErr handles error logging
func SystemErr(err error) error {
	if err == nil {
		return nil
	}
	gErr, ok := err.(errors.Error)
	if !ok {
		log.Error(err)
		return err
	}

	var entry = log.NewEntry(log.StandardLogger())
	ent := entry.WithFields(errFields(gErr))
	switch errors.Severity(gErr) {
	case log.WarnLevel:
		ent.Warnf("%v", err)
	case log.InfoLevel:
		ent.Infof("%v", err)
	case log.DebugLevel:
		ent.Debugf("%v", err)
	default:
		ent.Errorf("%v", err)
		return err
	}
	return nil
}

func errFields(err errors.Error) log.Fields {
	f := log.Fields{}
	f["kind"] = errors.KindText(err)
	var ops []string
	for _, op := range errors.Ops(err) {
		ops = append(ops, op.String())
	}
	f["ops"] = strings.Join(ops, " -> ")

	return f
}
