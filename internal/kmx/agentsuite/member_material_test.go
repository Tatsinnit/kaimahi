package agentsuite

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	agentsuite "github.com/agent-primitives/suite"
)

func TestProjectMemberBindsDescriptorsAndOwnsMaterial(t *testing.T) {
	files := readMaterialFixture(t, filepath.Join("..", "..", "..", "testdata", "agentsuite", "minimal"))
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	material, err := ProjectMember(t.Context(), files, digest, "writer")
	if err != nil {
		t.Fatal(err)
	}
	if material.ArtifactDigest != digest || material.SuiteName != "minimal" || material.MemberID != "writer" || material.GraphDigest == material.MemberDescriptor.Digest || material.MemberDescriptor.Digest == material.InstructionsDescriptor.Digest {
		t.Fatal("identity domains collapsed", material)
	}
	if string(material.Instructions) != "Write clearly.\n" || !reflect.DeepEqual(material.Model, agentsuite.ModelRequirement{Protocol: "openai-compatible", Model: "example-model"}) {
		t.Fatal("projection changed material", material)
	}
	material.Instructions[0] = 'X'
	material.Agent.Model.Model = "changed"
	fresh, err := ProjectMember(t.Context(), files, digest, "writer")
	if err != nil || string(fresh.Instructions) != "Write clearly.\n" || fresh.Model.Model != "example-model" {
		t.Fatal("returned material aliases retained input", err)
	}
}

func TestProjectMemberRefusesMismatchAndPreservesCancellation(t *testing.T) {
	base := readMaterialFixture(t, filepath.Join("..", "..", "..", "testdata", "agentsuite", "minimal"))
	digest := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, tc := range []struct {
		name   string
		mutate func(map[string][]byte)
		member string
		want   error
	}{
		{"unknown", func(map[string][]byte) {}, "absent", ErrMemberNotFound},
		{"member descriptor", func(files map[string][]byte) { files["agents/writer.json"][0] = 'X' }, "writer", ErrMemberMaterial},
		{"instructions descriptor", func(files map[string][]byte) { files["instructions/writer.md"][0] = 'X' }, "writer", ErrMemberMaterial},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := cloneMaterialFiles(base)
			tc.mutate(files)
			if _, err := ProjectMember(t.Context(), files, digest, tc.member); !errors.Is(err, tc.want) {
				t.Fatalf("wrong refusal: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ProjectMember(ctx, base, digest, "writer"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation identity lost: %v", err)
	}
}

func readMaterialFixture(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	if err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)], err = os.ReadFile(name)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

func cloneMaterialFiles(files map[string][]byte) map[string][]byte {
	clone := make(map[string][]byte, len(files))
	for name, body := range files {
		clone[name] = append([]byte(nil), body...)
	}
	return clone
}
