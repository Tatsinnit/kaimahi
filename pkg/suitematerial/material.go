package suitematerial

import (
	"bytes"
	"context"
	"errors"
	"io"

	agentsuite "github.com/agent-primitives/suite"
	kmxsuite "github.com/kaimahi-agents/kaimahi/internal/kmx/agentsuite"
	"github.com/kaimahi-agents/kaimahi/internal/kmx/agentsuite/materialio"
)

// Selection identifies one exact member. ArtifactDigest may be omitted only
// when the OCI index has one unambiguous root.
type Selection struct {
	ArtifactDigest string
	MemberID       string
}

// ArtifactIdentity is the exact raw OCI manifest descriptor.
type ArtifactIdentity struct {
	MediaType string
	Digest    string
	Size      int64
}

// SuiteIdentity distinguishes the logical graph from its enclosing OCI artifact.
type SuiteIdentity struct {
	Name        string
	GraphDigest string
}

// MemberIdentity identifies the selected canonical Agent document.
type MemberIdentity struct {
	ID         string
	Descriptor agentsuite.Descriptor
}

// Material is a caller-owned read-only projection, not an execution plan or
// evidence of runtime support, policy enforcement, readiness, or evaluation.
type Material struct {
	Artifact               ArtifactIdentity
	Suite                  SuiteIdentity
	Member                 MemberIdentity
	Description            string
	Instructions           []byte
	InstructionsDescriptor agentsuite.Descriptor
	Model                  agentsuite.ModelRequirement
}

// RefusalCode is a stable, payload-free refusal category.
type RefusalCode string

const (
	RefusalSelection         RefusalCode = "selection"
	RefusalArtifact          RefusalCode = "artifact"
	RefusalLimit             RefusalCode = "limit"
	RefusalAmbiguousArtifact RefusalCode = "ambiguous-artifact"
	RefusalArtifactNotFound  RefusalCode = "artifact-not-found"
	RefusalMemberNotFound    RefusalCode = "member-not-found"
	RefusalMaterial          RefusalCode = "material"
	RefusalCanceled          RefusalCode = "canceled"
)

// Refusal contains no untrusted paths, document values, or parser diagnostics.
type Refusal struct {
	Code  RefusalCode
	cause error
}

func (r *Refusal) Error() string {
	if r == nil {
		return "suite material loading refused"
	}
	return "suite material loading refused: " + string(r.Code)
}

func (r *Refusal) Unwrap() error {
	if r == nil {
		return nil
	}
	return r.cause
}

// LoadLayout loads from a local OCI image-layout directory without mutating it.
func LoadLayout(ctx context.Context, layout string, selection Selection) (Material, error) {
	if err := validateSelection(ctx, selection); err != nil {
		return Material{}, err
	}
	artifact, err := materialio.ReadLayout(ctx, layout, selection.ArtifactDigest)
	if err != nil {
		return Material{}, refusalFor(ctx, err)
	}
	return project(ctx, artifact, selection.MemberID)
}

// LoadArchive loads an uncompressed OCI image-layout tar from an already-open
// ReaderAt. At most 64 MiB is frozen before validation and projection.
func LoadArchive(ctx context.Context, source io.ReaderAt, size int64, selection Selection) (Material, error) {
	if err := validateSelection(ctx, selection); err != nil {
		return Material{}, err
	}
	artifact, err := materialio.ReadArchive(ctx, source, size, selection.ArtifactDigest)
	if err != nil {
		return Material{}, refusalFor(ctx, err)
	}
	return project(ctx, artifact, selection.MemberID)
}

func validateSelection(ctx context.Context, selection Selection) error {
	if err := ctx.Err(); err != nil {
		return &Refusal{Code: RefusalCanceled, cause: err}
	}
	if selection.MemberID == "" || selection.ArtifactDigest != "" && !agentsuite.IsValidDigest(selection.ArtifactDigest) {
		return &Refusal{Code: RefusalSelection}
	}
	return nil
}

func project(ctx context.Context, artifact *materialio.Artifact, memberID string) (Material, error) {
	if artifact == nil {
		return Material{}, &Refusal{Code: RefusalArtifact}
	}
	descriptor := artifact.Descriptor()
	projection, err := kmxsuite.ProjectMember(ctx, artifact.Files(), descriptor.Digest.String(), memberID)
	if err != nil {
		return Material{}, refusalFor(ctx, err)
	}
	return Material{
		Artifact:               ArtifactIdentity{MediaType: descriptor.MediaType, Digest: descriptor.Digest.String(), Size: descriptor.Size},
		Suite:                  SuiteIdentity{Name: projection.SuiteName, GraphDigest: projection.GraphDigest},
		Member:                 MemberIdentity{ID: projection.MemberID, Descriptor: cloneDescriptor(projection.MemberDescriptor)},
		Description:            projection.Description,
		Instructions:           bytes.Clone(projection.Instructions),
		InstructionsDescriptor: cloneDescriptor(projection.InstructionsDescriptor),
		Model:                  cloneModel(projection.Model),
	}, nil
}

func refusalFor(ctx context.Context, err error) error {
	if cause := ctx.Err(); cause != nil {
		return &Refusal{Code: RefusalCanceled, cause: cause}
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, cause) {
			return &Refusal{Code: RefusalCanceled, cause: cause}
		}
	}
	switch {
	case errors.Is(err, materialio.ErrArtifactLimit):
		return &Refusal{Code: RefusalLimit}
	case errors.Is(err, materialio.ErrArtifactAmbiguous):
		return &Refusal{Code: RefusalAmbiguousArtifact}
	case errors.Is(err, materialio.ErrArtifactNotFound):
		return &Refusal{Code: RefusalArtifactNotFound}
	case errors.Is(err, kmxsuite.ErrMemberNotFound):
		return &Refusal{Code: RefusalMemberNotFound}
	case errors.Is(err, kmxsuite.ErrMemberMaterial):
		return &Refusal{Code: RefusalMaterial}
	default:
		return &Refusal{Code: RefusalArtifact}
	}
}

func cloneDescriptor(value agentsuite.Descriptor) agentsuite.Descriptor {
	if value.Annotations != nil {
		value.Annotations = make(map[string]string, len(value.Annotations))
		for key, item := range value.Annotations {
			value.Annotations[key] = item
		}
	}
	return value
}

func cloneModel(value agentsuite.ModelRequirement) agentsuite.ModelRequirement {
	if value.Capabilities != nil {
		capabilities := *value.Capabilities
		value.Capabilities = &capabilities
	}
	value.SecretRefs = append([]string(nil), value.SecretRefs...)
	return value
}
