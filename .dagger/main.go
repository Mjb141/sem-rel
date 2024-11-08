// A generated module for SemRel functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/sem-rel/internal/dagger"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"
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

// Returns a container that echoes whatever string argument is provided
func (m *SemRel) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *SemRel) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}

func (m *SemRel) Release(
	ctx context.Context,
	// +defaultPath="/.releaserc.json"
	releaserc *dagger.File,
	// +defaultPath="/"
	dir *dagger.Directory,
	// +default="Github"
	provider string,
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
		log.Panic("Failed to read .releaserc.json contents", err)
	}

	var config Config
	if err := json.Unmarshal([]byte(rules), &config); err != nil {
		log.Panic("Failed to Unmarshal .releaserc.json", err)
	}

	currentBranch, err := dag.GitInfo(dir).Branch(ctx)
	if err != nil {
		log.Panic("Could not get the current branch name", err)
	}

	var branchNamesInReleaseRc []string
	for _, branch := range config.Branches {
		fmt.Println(fmt.Sprintf("Branch found: %s", branch.Name))
		branchNamesInReleaseRc = append(branchNamesInReleaseRc, branch.Name)
	}

	currentBranchInReleaseBranches := slices.Contains(branchNamesInReleaseRc, currentBranch)
	if !currentBranchInReleaseBranches {
		config.Branches = append(config.Branches, Branch{currentBranch, false, ""})
	}

	updatedConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	currentTime := time.Now().Format("11:00:00")
	ctr := dag.Container().
		From("hoppr/semantic-release").
		WithEnvVariable("BUST_CACHE", currentTime).
		WithEnvVariable("BRANCH_NAME", currentBranch).
		WithEnvVariable("BRANCHES_IN_RELEASERC", strings.Join(branchNamesInReleaseRc, ", ")).
		WithEnvVariable("CURRENT_BRANCH_IN_RELEASERC", fmt.Sprint(currentBranchInReleaseBranches)).
		WithSecretVariable(tokenKey, token).
		WithDirectory("/src", dir).
		WithWorkdir("/src").
		WithoutFile(".releaserc.json").
		WithNewFile(".releaserc.json", string(updatedConfig)).
		WithExec([]string{"cat", ".releaserc.json"}).
		WithExec([]string{"semantic-release", "--no-ci", "--dry-run"})

	return ctr, err
}
