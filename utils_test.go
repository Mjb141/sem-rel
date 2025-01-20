package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (m *SemRel) TestSemanticReleaseCommandLocalDryRun(t *testing.T) {
	dryRun := true
	checkIfCI := false
	ctx := context.Background()
	semanticReleaseCommand := m.SemanticReleaseCommand(ctx, dryRun, checkIfCI)

	assert.Contains(t, semanticReleaseCommand, "--dry-run")
	assert.Contains(t, semanticReleaseCommand, "--no-ci")
}

func (m *SemRel) TestSemanticReleaseCommandLocalRelease(t *testing.T) {
	dryRun := false
	checkIfCI := false
	ctx := context.Background()
	semanticReleaseCommand := m.SemanticReleaseCommand(ctx, dryRun, checkIfCI)

	assert.Equal(t, []string{"semantic-release", "--no-ci"}, semanticReleaseCommand)
}
