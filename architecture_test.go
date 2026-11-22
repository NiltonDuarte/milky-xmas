package milkyxmas_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestFCISDependencies(t *testing.T) {
	cmd := exec.CommandContext(t.Context(), "go", "list", "-json=ImportPath,Module,Deps", "./...")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("go list: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("go list: %v", err)
	}

	packages := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg struct {
			ImportPath string
			Module     struct{ Path string }
			Deps       []string
		}
		err := packages.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decode go list output: %v", err)
		}

		modulePrefix := pkg.Module.Path + "/"
		packagePath := strings.TrimPrefix(pkg.ImportPath, modulePrefix)
		isCore := inPackageTree(packagePath, "internal/core")
		isApp := inPackageTree(packagePath, "internal/app")
		if !isCore && !isApp {
			continue
		}

		for _, dep := range pkg.Deps {
			dependencyPath, local := strings.CutPrefix(dep, modulePrefix)
			if !local && dep != pkg.Module.Path {
				continue
			}
			switch {
			case isCore && !inPackageTree(dependencyPath, "internal/core"):
				t.Errorf("%s depends on %s outside the functional core", pkg.ImportPath, dep)
			case isApp && (inPackageTree(dependencyPath, "cmd") || inPackageTree(dependencyPath, "internal/adapter")):
				t.Errorf("%s depends on %s outside the shared application layer", pkg.ImportPath, dep)
			}
		}
	}
}

func inPackageTree(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+"/")
}
