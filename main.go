// Configurable Semantic Release
//
// Configurable Semantic Release module. Options are provided to modify your
// releaserc file on demand for testing purposes (e.g. local/dry/no-ci runs,
// unlisted branches etc).
//
// Usable with Github or Gitlab with a PAT token.

package main

import (
	"context"
	"dagger/sem-rel/internal/dagger"
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	SemRelImage = "hoppr/semantic-release@sha256:64cb33458281ab15a9249747c74d498b54d2ea125047c4fd1f24b3f04b28bf00"
)

type SemRel struct {
	// Allow the user to add the current branch to .releaserc.json
	// +private
	AddCurrentBranch bool
	// All the user to remove @semantic-release/github or @semantic-release/gitlab from plugins
	// +private
	RemoveGitProvider bool
	// Semantic Release --dry-run option
	// +private
	DryRun bool
	// Semantic Release --no-ci option
	// +private
	CheckIfCi bool
}

// PluginOptions defines the interface for plugin configuration options
type PluginOptions interface {
	DaggerObject
}

// Config represents the overall JSON structure
type Config struct {
	Branches []Branch        `json:"branches"`
	Plugins  []PluginElement `json:"plugins"`
}

// Branch represents a branch configuration
type Branch struct {
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease,omitempty"`
	Channel    string `json:"channel,omitempty"`
}

// Plugin represents a plugin in the JSON data
type PluginElement struct {
	Name    string
	Options map[string]interface{}
}

func (p *PluginElement) UnmarshalJSON(data []byte) error {
	var nameOnly string
	if err := json.Unmarshal(data, &nameOnly); err == nil {
		p.Name = nameOnly
		return nil
	}

	var nameWithOptions []interface{}
	if err := json.Unmarshal(data, &nameWithOptions); err != nil {
		return err
	}

	if len(nameWithOptions) > 0 {
		if name, ok := nameWithOptions[0].(string); ok {
			p.Name = name
		}
	}

	if len(nameWithOptions) > 1 {
		if options, ok := nameWithOptions[1].(map[string]interface{}); ok {
			p.Options = options
		}
	}

	return nil
}

// Configure Semantic Release
func (m *SemRel) Configure(
	ctx context.Context,
	// Add the current branch to the 'branches' key in your .releaserc file
	// +default=false
	addCurrentBranch bool,
	// Remove the Github/Gitlab from plugins (if you do not have a token)
	// +default=false
	removeGitProvider bool,
	// The Semantic Release --dry-run flag for testing
	// +default=true
	dryRun bool,
	// The Semantic Release --check-if-ci flag for local execution
	// +default=false
	checkIfCi bool,
) *SemRel {
	m.AddCurrentBranch = addCurrentBranch
	m.RemoveGitProvider = removeGitProvider
	m.DryRun = dryRun
	m.CheckIfCi = checkIfCi
	return m
}

// Run Semantic Release
func (m *SemRel) Release(
	ctx context.Context,
	// .releaserc (or equivalent), defaults to .releaserc.json
	// +defaultPath="/.releaserc.json"
	releaserc *dagger.File,
	// Directory containing app that should be assessed for release
	// +defaultPath="/"
	dir *dagger.Directory,
	// Git provider, Github or Gitlab. Determines if GH_TOKEN or GL_TOKEN is provided to Semantic Release
	// +default="Github"
	provider string,
	// PAT for either Github or Gitlab. Will be provided to Semantic Release as GH_TOKEN or GL_TOKEN
	token *dagger.Secret,
) (*dagger.Container, error) {
	var tokenKey string
	if provider == "Github" {
		tokenKey = "GH_TOKEN"
	} else {
		tokenKey = "GL_TOKEN"
	}

	rules, err := releaserc.Contents(ctx)
	if err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to read .releaserc.json contents")

		return nil, err
	}

	var config Config
	if err := json.Unmarshal([]byte(rules), &config); err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to Unmarshal .releaserc.json")

		return nil, err
	}

	// Modify configuration if required
	if m.AddCurrentBranch {
		log.
			Debug().
			Bool("addCurrentBranch", m.AddCurrentBranch).
			Msg("Attempting to add %s to config.Branches")

		branches, err := m.AddBranchToReleaseRc(ctx, dir, config.Branches)
		if err != nil {
			log.
				Error().
				Err(err).
				Msg("Failed to add branch to '.releaserc.json'")

			return nil, err
		}

		config.Branches = branches
	}

	if m.RemoveGitProvider {
		log.
			Debug().
			Bool("removeGitProvider", m.RemoveGitProvider).
			Msg("Attempting to remove @semantic-release/github or @semantic-release/gitlab from plugins")

		// config.Plugins = m.RemoveGitPlugin(ctx, dir, config.Plugins)
	}

	// Modify command if required
	semanticReleaseCommand := m.SemanticReleaseCommand(ctx, m.DryRun, m.CheckIfCi)

	updatedConfig, err := json.Marshal(config)
	if err != nil {
		log.
			Error().
			Err(err).
			Msg("Failed to recreate .releaserc.json from config")

		return nil, err
	}

	currentTime := time.Now().Format("11:00:00")
	ctr := dag.Container().
		From(SemRelImage).
		WithEnvVariable("BUST_CACHE", currentTime).
		WithEnvVariable("SEMREL_COMMAND", strings.Join(semanticReleaseCommand, " ")).
		WithSecretVariable(tokenKey, token).
		WithDirectory("/src", dir).
		WithWorkdir("/src").
		WithoutFile(".releaserc.json").
		WithNewFile(".releaserc.json", string(updatedConfig)).
		WithExec([]string{"cat", ".releaserc.json"}).
		WithExec(semanticReleaseCommand)

	return ctr, err
}
