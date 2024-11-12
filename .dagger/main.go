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
	"fmt"
	"github.com/rs/zerolog/log"
	"slices"
	"time"
)

type SemRel struct{}

type Branch struct {
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease,omitempty"`
	Channel    string `json:"channel,omitempty"`
}

// Plugin represents a plugin in the JSON data
type Plugin struct {
	Name    string            `json:"name"`
	Options map[string]string `json:"options,omitempty"`
}

// Config represents the overall JSON structure
type Config struct {
	Branches []Branch    `json:"branches"`
	Plugins  interface{} `json:"plugins"`
}

func (m *SemRel) AddBranchToReleaseRc(ctx context.Context, dir *dagger.Directory, branches []Branch) ([]Branch, error) {
	currentBranch, err := dag.GitInfo(dir).Branch(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Could not get the current branch name")
		return branches, err
	}

	var branchNamesInReleaseRc []string
	for _, branch := range branches {
		log.Info().Str("Branch", branch.Name).Msg("Branch found")
		branchNamesInReleaseRc = append(branchNamesInReleaseRc, branch.Name)
	}

	currentBranchInReleaseBranches := slices.Contains(branchNamesInReleaseRc, currentBranch)

	if currentBranchInReleaseBranches {
		log.Info().Msg(fmt.Sprintf("Branch %s already found in config.Branches", currentBranch))
		return branches, nil
	}

	log.Info().Msg(fmt.Sprintf("Adding branch %s to config.Branches", currentBranch))
	return append(branches, Branch{currentBranch, false, ""}), nil
}

func (m *SemRel) Release(
	ctx context.Context,
	// +defaultPath="/.releaserc.json"
	releaserc *dagger.File,
	// +defaultPath="/"
	dir *dagger.Directory,
	// +default="Github"
	provider string,
	// +default=false
	addCurrentBranch bool,
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

	if addCurrentBranch {
		log.Debug().Bool("addCurrentBranch", addCurrentBranch).Msg("Attempting to add %s to config.Branches")
		branches, err := m.AddBranchToReleaseRc(ctx, dir, config.Branches)
		if err != nil {
			return nil, err
		}

		config.Branches = branches
	}

	updatedConfig, err := json.Marshal(config)
	if err != nil {
		log.Error().Err(err).Msg("Failed to recreate .releaserc.json from config")
		return nil, err
	}

	currentTime := time.Now().Format("11:00:00")
	ctr := dag.Container().
		From("hoppr/semantic-release").
		WithEnvVariable("BUST_CACHE", currentTime).
		WithSecretVariable(tokenKey, token).
		WithDirectory("/src", dir).
		WithWorkdir("/src").
		WithoutFile(".releaserc.json").
		WithNewFile(".releaserc.json", string(updatedConfig)).
		WithExec([]string{"cat", ".releaserc.json"}).
		WithExec([]string{"semantic-release", "--no-ci", "--dry-run"})

	return ctr, err
}
