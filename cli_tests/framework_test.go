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
	frameworkCliName  = "myframework"
	frameworkYamlName = "myframework-yaml"
)

func TestFrameworkHelp(t *testing.T) {
	b, err := exec.Command(ketch, "framework", "--help").CombinedOutput()
	require.Nil(t, err)
	require.Contains(t, string(b), "add")
	require.Contains(t, string(b), "export")
	require.Contains(t, string(b), "list")
	require.Contains(t, string(b), "remove")
	require.Contains(t, string(b), "update")
}

func TestFrameworkByCLI(t *testing.T) {
	defer cleanupFramework(frameworkCliName)

	// add framework
	b, err := exec.Command(ketch, "framework", "add", frameworkCliName, "--ingress-service-endpoint", ingress, "--ingress-type", "traefik").CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully added!")

	// list framework
	b, err = exec.Command(ketch, "framework", "list").CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile("NAME[ \t]+STATUS[ \t]+NAMESPACE[ \t]+INGRESS TYPE[ \t]+INGRESS CLASS NAME[ \t]+CLUSTER ISSUER[ \t]+APPS").Match(b), string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("%s[ \t]+[Created \t]+ketch-%s[ \t]+traefik[ \t]+traefik", frameworkCliName, frameworkCliName)).Match(b), string(b))

	// update framework
	b, err = exec.Command(ketch, "framework", "update", frameworkCliName, "--app-quota-limit", "2").CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully updated!", string(b))

	// export framework
	b, err = exec.Command(ketch, "framework", "export", frameworkCliName).CombinedOutput()
	require.Nil(t, err, string(b))
	// FIXME: the quotes around serviceEndpoint are incorrect
	require.True(t, regexp.MustCompile(fmt.Sprintf("appQuotaLimit: 2\ningressController:\n  className: traefik\n(  serviceEndpoint: '''.*'''\n)?  type: traefik\nname: %s\nnamespace: ketch-%s", frameworkCliName, frameworkCliName)).Match(b), string(b))

	// remove framework
	cmd := exec.Command(ketch, "framework", "remove", frameworkCliName)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("ketch-%s", frameworkCliName))
	cmd.Stdin = &buf
	b, err = cmd.CombinedOutput()
	require.Nil(t, err)
	require.Contains(t, string(b), "Framework successfully removed!", string(b))
}

func TestFrameworkByYaml(t *testing.T) {
	defer cleanupFramework(frameworkYamlName)

	// add framework
	temp, err := os.CreateTemp(t.TempDir(), "*.yaml")
	require.Nil(t, err)
	defer os.Remove(temp.Name())
	_, err = temp.WriteString(fmt.Sprintf(`name: %s
app-quota-limit: 1
ingressController:
 className: traefik
 serviceEndpoint: %s
 type: traefik`, frameworkYamlName, ingress))

	b, err := exec.Command(ketch, "framework", "add", temp.Name()).CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully added!")

	// list framework
	b, err = exec.Command(ketch, "framework", "list").CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("%s[ \t]+[Created \t]+ketch-%s[ \t]+traefik[ \t]+traefik", frameworkYamlName, frameworkYamlName)).Match(b), string(b))

	// update framework
	b, err = exec.Command(ketch, "framework", "update", frameworkYamlName, "--app-quota-limit", "2").CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "Successfully updated!", string(b))

	// export framework
	b, err = exec.Command(ketch, "framework", "export", frameworkYamlName).CombinedOutput()
	require.Nil(t, err, string(b))
	require.True(t, regexp.MustCompile(fmt.Sprintf("appQuotaLimit: 2\ningressController:\n  className: traefik\n(  serviceEndpoint: .*\n)?  type: traefik\nname: %s\nnamespace: ketch-%s\nversion: v1", frameworkYamlName, frameworkYamlName)).Match(b), string(b))

	// remove framework
	cmd := exec.Command(ketch, "framework", "remove", frameworkYamlName)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("ketch-%s", frameworkYamlName))
	cmd.Stdin = &buf
	b, err = cmd.CombinedOutput()
	require.Nil(t, err)
	require.Contains(t, string(b), "Framework successfully removed!", string(b))
}
