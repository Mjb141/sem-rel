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

type SemRel struct {
	// Allow the user to add the current branch to .releaserc.json
	// +private
	AddCurrentBranch bool
	// All the user to remove @semantic-release/github or @semantic-release/gitlab from plugins
	// +private
	removeGitProvider bool
	// Semantic Release --dry-run option
	// +private
	DryRun bool
	// Semantic Release --no-ci option
	// +private
	CheckIfCi bool
}

type Branch struct {
	Name       string `json:"name"`
	Channel    string `json:"channel,omitempty"`
	Prerelease bool   `json:"prerelease,omitempty"`
}

// Plugin represents a plugin in the JSON data
type Plugin struct {
	Options map[string]string `json:"options,omitempty"`
	Name    string            `json:"name"`
}

// Config represents the overall JSON structure
type Config struct {
	Branches []Branch      `json:"branches"`
	Plugins  []interface{} `json:"plugins"`
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
	m.removeGitProvider = removeGitProvider
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
		log.Error().Err(err).Msg("Failed to read .releaserc.json contents")
	}

	var config Config
	if err := json.Unmarshal([]byte(rules), &config); err != nil {
		log.Error().Err(err).Msg("Failed to Unmarshal .releaserc.json")
	}

	log.Debug().Bool("add-current-branch", m.AddCurrentBranch).Msg("Configuration")
	// Modify configuration if required
	if m.AddCurrentBranch {
		log.
			Debug().
			Bool("addCurrentBranch", m.AddCurrentBranch).
			Msg("Attempting to add %s to config.Branches")

		branches, err := AddBranchToReleaseRc(ctx, dir, config.Branches)
		if err != nil {
			return nil, err
		}

		config.Branches = branches
	}

	log.Debug().Bool("remove-git-provider", m.removeGitProvider).Msg("Configuration")
	if m.removeGitProvider {
		log.
			Debug().
			Bool("removeGitProvider", m.removeGitProvider).
			Msg("Attempting to remove @semantic-release/{github/gitlab} from plugins")

		RemoveGitProvider(ctx, dir, config.Plugins)
	}

	// Modify command if required
	semanticReleaseCommand := SemanticReleaseCommand(ctx, m.DryRun, m.CheckIfCi)

	updatedConfig, err := json.Marshal(config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to recreate .releaserc.json from config")
		return nil, err
	}

	currentTime := time.Now().Format("11:00:00")
	ctr := dag.Container().
		From("hoppr/semantic-release").
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
