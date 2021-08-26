// +build integration

package cli_tests

import (
	"os/exec"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilderList(t *testing.T) {
	b, err := exec.Command(ketch, "builder", "list").CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile("VENDOR[ \t]+IMAGE[ \t]+DESCRIPTION").Match(b))
	require.True(t, regexp.MustCompile("Google[ \t]+gcr.io/buildpacks/builder:v1[ \t]+GCP Builder for all runtimes").Match(b))
}
