//go:build cli_integration

package integration

import (
	"os"
	"testing"
)

var (
	shen    string
	shenctl string
)

func TestMain(m *testing.M) {
	shen = compileBinary("../../cmd/shen", "shen")
	shenctl = compileBinary("../../cli/shenctl", "shenctl")

	exitCode := m.Run()

	os.Remove(shen)
	os.Remove(shenctl)
	os.Exit(exitCode)
}
