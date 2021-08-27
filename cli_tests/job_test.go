// +build integration

package cli_tests

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

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
	defer func() {
		cleanupJob(jobName)
		cleanupFramework(jobFramework)
	}()

	// add framework
	b, err := exec.Command(ketch, "framework", "add", jobFramework).CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully added!")
	err = retry(ketch, []string{"framework", "list"}, "", jobFramework, 3, 3)
	require.Nil(t, err)
	time.Sleep(time.Second * 3) // TODO - job test fails if we don't wait even though we assure the framework is created above

	// add job
	temp, err := os.CreateTemp(t.TempDir(), "*.yaml")
	require.Nil(t, err)
	defer os.Remove(temp.Name())
	_, err = temp.WriteString(fmt.Sprintf(`name: %s
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
}
