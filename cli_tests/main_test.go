// +build integration

package cli_tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	ketch   string // ketch executable path
	ingress string // ingress IP
)

func init() {
	// set ingress IP
	b, err := exec.Command("kubectl", "get", "svc", "traefik", "-o", "jsonpath='{.status.loadBalancer.ingress[0].ip}'").Output()
	if err != nil {
		panic(err)
	}
	ingress = string(b)

	// set ketch executable path
	ketchExecPath := os.Getenv("KETCH_EXECUTABLE_PATH")
	if ketchExecPath != "" {
		ketch = ketchExecPath
		return
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	ketch = filepath.Join(pwd, "bin", "ketch")
}

// retry tries a command <times> in intervals of <wait> seconds.
// If <match> is never found in command output, an error is returned containing
// all aggregated output.
func retry(name string, args []string, input string, match string, times, wait int) error {
	sb := strings.Builder{}
	for i := 0; i < times; i++ {
		cmd := exec.Command(name, args...)
		if input != "" {
			var buf bytes.Buffer
			buf.WriteString(input)
			cmd.Stdin = &buf
		}
		b, _ := cmd.CombinedOutput() // sometimes we want exit status 1
		sb.Write(b)
		sb.WriteString("\n")

		if strings.Contains(string(b), match) {
			return nil
		}
		if i < times-1 {
			fmt.Println("retrying command: ", name, args)
			time.Sleep(time.Second * time.Duration(wait))
		}
	}
	return fmt.Errorf("retry failed on command %s. Output: %s", name, sb.String())
}

// cleanupFramework attempts to remove a framework. Can be called in a defer to
// assure framework cleanup in the event of a failed test
func cleanupFramework(name string) {
	cmd := exec.Command(ketch, "framework", "remove", name)
	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("ketch-%s", name)))
	cmd.Stdin = &buf
	cmd.Run()
}

// cleanupApp attempts to remove an app. Can be called in a defer to
// assure app cleanup in the event of a failed test
func cleanupApp(name string) {
	cmd := exec.Command(ketch, "app", "remove", name)
	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("ketch-%s", name)))
	cmd.Stdin = &buf
	cmd.Run()
}

func TestHelp(t *testing.T) {
	b, err := exec.Command(ketch, "help").CombinedOutput()
	require.Nil(t, err, string(b))
	require.Contains(t, string(b), "For details see https://theketch.io")
	require.Contains(t, string(b), "Available Commands")
	require.Contains(t, string(b), "Flags")
}
