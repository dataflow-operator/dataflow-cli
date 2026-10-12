package scripts

import (
	"os/exec"
	"strings"
	"testing"
)

func TestHomebrewFormula(t *testing.T) {
	version := "1.4.2"
	darwinArm := strings.Repeat("ab", 32)
	darwinAmd := strings.Repeat("cd", 32)
	linuxArm := strings.Repeat("ef", 32)
	linuxAmd := strings.Repeat("12", 32)

	out := renderFormula(t, version, darwinArm, darwinAmd, linuxArm, linuxAmd)

	base := "https://github.com/dataflow-operator/dataflow-cli/releases/download/v" + version
	checks := []string{
		`version "` + version + `"`,
		`license "Apache-2.0"`,
		`url "` + base + `/dataflow-cli-darwin-arm64.tar.gz"`,
		`sha256 "` + darwinArm + `"`,
		`url "` + base + `/dataflow-cli-darwin-amd64.tar.gz"`,
		`sha256 "` + darwinAmd + `"`,
		`url "` + base + `/dataflow-cli-linux-arm64.tar.gz"`,
		`sha256 "` + linuxArm + `"`,
		`url "` + base + `/dataflow-cli-linux-amd64.tar.gz"`,
		`sha256 "` + linuxAmd + `"`,
		`bin.install "dataflow-cli"`,
		`shell_output("#{bin}/dataflow-cli --version")`,
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("formula missing %q\n%s", want, out)
		}
	}
}

func TestHomebrewFormulaRejectsBadInput(t *testing.T) {
	sha := strings.Repeat("ab", 32)
	if err := execFormula("v1.4.2", sha, sha, sha, sha); err == nil {
		t.Fatal("version with leading v was accepted")
	}
	if err := execFormula("1.4.2", "abc", sha, sha, sha); err == nil {
		t.Fatal("short sha256 was accepted")
	}
}

func renderFormula(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("sh", append([]string{"homebrew-formula.sh"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("homebrew-formula.sh: %v", err)
	}
	return string(out)
}

func execFormula(args ...string) error {
	cmd := exec.Command("sh", append([]string{"homebrew-formula.sh"}, args...)...)
	return cmd.Run()
}
