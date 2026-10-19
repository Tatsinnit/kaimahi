package suitematerial_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPublicPackageKeepsOperationalDependenciesOut(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", ".").CombinedOutput()
	if err != nil {
		t.Fatal(err, string(output))
	}
	for _, dependency := range strings.Fields(string(output)) {
		for _, forbidden := range []string{
			"os/exec",
			"github.com/kaimahi-agents/kaimahi/cmd/",
			"github.com/kaimahi-agents/kaimahi/internal/kmx/app",
			"github.com/kaimahi-agents/kaimahi/internal/kmx/agentsuite/oras",
			"github.com/kaimahi-agents/kaimahi/internal/kmx/agentsuite/workspace",
			"github.com/kaimahi-agents/kaimahi/internal/kmx/config",
			"github.com/kaimahi-agents/kaimahi/internal/kmx/runtime",
			"github.com/spf13/",
			"k8s.io/",
			"oras.land/oras-go/v2/registry",
		} {
			if dependency == forbidden || strings.HasPrefix(dependency, forbidden) {
				t.Errorf("public read-only loader depends on operational package %q", dependency)
			}
		}
	}
}
