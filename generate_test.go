package yamlnetlink_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/yamlnetlink"
	"golang.org/x/exp/slices"
)

func TestGenerate(t *testing.T) {
	// Generate a nlctrl client in another package which we can compile, run,
	// and verify it works.
	code, err := yamlnetlink.Generate(nlctrl(), &yamlnetlink.Config{Package: "main"})
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	if err := os.Remove("testdata/nlctrl.go"); err != nil {
		t.Fatalf("failed to remove generate code: %v", err)
	}

	if err := os.WriteFile("testdata/nlctrl.go", code, 0o644); err != nil {
		t.Fatalf("failed to write generated code: %v", err)
	}

	if err := os.Chdir("./testdata"); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// The output is a known JSON format we can compare.
	out, err := exec.Command("go", "run", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to execute code: %v\noutput: %q", err, out)
	}

	type family struct {
		ID   uint16 `json:"FamilyId"`
		Name string `json:"FamilyName"`
	}

	type stdout struct {
		Family   family   `json:"family"`
		Families []string `json:"families"`
	}

	var got stdout
	if err := json.Unmarshal(out, &got); err != nil {
		t.Logf("stdout: %s", string(out))
		t.Fatalf("failed to unmarshal family: %v", err)
	}

	// nlctrl always occupies the same IDs.
	if diff := cmp.Diff(family{ID: 16, Name: "nlctrl"}, got.Family); diff != "" {
		t.Fatalf("unexpected generic netlink family (-want +got):\n%s", diff)
	}

	// Expect to find nlctrl in the dump list.
	if !slices.Contains(got.Families, "nlctrl") {
		t.Fatalf("did not find nlctrl: %v", got.Families)
	}
}
