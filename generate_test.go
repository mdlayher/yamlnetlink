package yamlnetlink_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/yamlnetlink"
	"golang.org/x/exp/slices"
)

func TestGenerateNlctrl(t *testing.T) {
	out := generate(t, "nlctrl")

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

func TestGenerateEthtool(t *testing.T) {
	_ = generate(t, "ethtool")
	// TODO!
}

// generate generates and executes Go code for the specified family using the
// family's directory under testdata.
func generate(t *testing.T, family string) []byte {
	t.Helper()

	// Example: ./testdata/nlctrl/nlctrl
	path := filepath.Join("testdata", family, family)
	f, err := os.Open(path + ".yaml")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	spec, err := yamlnetlink.Parse(f)
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}
	_ = f.Close()

	// Generate a client in another package which we can compile, run, and
	// verify it works.
	code, err := yamlnetlink.Generate(spec, &yamlnetlink.Config{Package: "main"})
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	_ = os.Remove(path + ".go")
	if err := os.WriteFile(path+".go", code, 0o644); err != nil {
		t.Fatalf("failed to write generated code: %v", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Dir(path)); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(pwd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()

	// Execute the code and pass stdout back to the caller.
	out, err := exec.Command("go", "run", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("failed to execute code: %v\noutput: %q", err, out)
	}

	return out
}
