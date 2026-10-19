package workspace

import (
	"context"
	"errors"
	"strings"
	"testing"

	agentsuite "github.com/agent-primitives/suite"
)

func TestLoadNativeMemberPreservesCancellationFromPublicProjection(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	suite := agentsuite.Suite{Agents: []agentsuite.ManifestRef{{ID: "writer"}}}
	_, err := loadNativeMember(ctx, map[string][]byte{}, suite, nativeRecord{}, RuntimeProfile{}, "sha256:"+strings.Repeat("a", 64), "", 0)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation identity lost: %v", err)
	}
}
