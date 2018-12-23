package deprecate

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update .golden files")

func TestNotice(t *testing.T) {
	f, err := ioutil.TempFile("", "output.txt")
	require.NoError(t, err)

	color.NoColor = true

	log.Info("first")
	Notice("foo.bar.whatever")
	log.Info("last")

	require.NoError(t, f.Close())

	bts, err := ioutil.ReadFile(f.Name())
	require.NoError(t, err)

	const golden = "testdata/output.txt.golden"
	if *update {
		require.NoError(t, ioutil.WriteFile(golden, bts, 0655))
	}

	gbts, err := ioutil.ReadFile(golden)
	require.NoError(t, err)

	require.Equal(t, string(gbts), string(bts))
}
