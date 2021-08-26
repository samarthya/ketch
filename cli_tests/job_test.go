// +build integration

package cli_tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	jobName      = "myjob"
	jobFramework = "jobframework"
)

func TestJobHelp(t *testing.T) {
	b, err := exec.Command(ketch, "job", "--help").CombinedOutput()
	require.Nil(t, err)
	require.Contains(t, string(b), "deploy")
	require.Contains(t, string(b), "export")
	require.Contains(t, string(b), "list")
	require.Contains(t, string(b), "remove")
}

func TestJobByYaml(t *testing.T) {
	defer cleanupFramework(jobFramework)

	// add framework
	b, err := exec.Command(ketch, "framework", "add", jobFramework).CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully added!")

	err = retry(ketch, []string{"framework", "list"}, "", "Created", 10, 5)
	require.Nil(t, err)

	// add job
	temp, err := os.CreateTemp(t.TempDir(), "*.yaml")
	require.Nil(t, err)
	defer os.Remove(temp.Name())
	temp.WriteString(fmt.Sprintf(`name: %s
version: v1
type: Job
framework: %s
description: "cli test job"
containers:
  - name: pi
    image: perl
    command:
      - "perl"
      - "-Mbignum=bpi"
      - "-wle"
      - "print bpi(2000)"`, jobName, jobFramework))

	b, err = exec.Command(ketch, "job", "deploy", temp.Name()).CombinedOutput()
	require.Nil(t, err, string(b))
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully added!")

	// assert job via kubectl
	err = retry("kubectl", []string{"get", "jobs", "-n", fmt.Sprintf("ketch-%s", jobFramework)}, "", "/1", 10, 4)
	require.Nil(t, err)

	// list job
	b, err = exec.Command(ketch, "job", "list").CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("%s[ \t]+[ \t]+v1[ \t]+%s[ \t]+cli test job", jobName, jobFramework)).Match(b), string(b))

	// export job
	b, err = exec.Command(ketch, "job", "export", jobName).CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("name: %s", jobName)).Match(b), string(b))
	require.True(t, regexp.MustCompile("version: v1").Match(b), string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("framework: %s", jobFramework)).Match(b), string(b))

	// delete job
	b, err = exec.Command(ketch, "job", "remove", jobName).CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully removed!")

	//remove framework
	cmd := exec.Command(ketch, "framework", "remove", jobFramework)
	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("ketch-%s", jobFramework)))
	cmd.Stdin = &buf
	b, err = cmd.CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Framework successfully removed!")
}
