package main

import (
	"context"
	"dagger/sem-rel/internal/dagger"
	"fmt"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

func AddBranchToReleaseRc(ctx context.Context, dir *dagger.Directory, branches []Branch) ([]Branch, error) {
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

func SemanticReleaseCommand(ctx context.Context, dryRun, checkIfCi bool) []string {
	cmd := []string{"semantic-release"}

	if dryRun {
		log.Debug().Str("Added", "--dry-run").Msg("Added option to 'semantic-release' command")
		cmd = append(cmd, "--dry-run")
	}

	if !checkIfCi {
		log.Debug().Str("Added", "--no-ci").Msg("Added option to 'semantic-release' command")
		cmd = append(cmd, "--no-ci")
	}

	log.Debug().Msg(fmt.Sprintf("Returning 'semantic-release' command: %s", strings.Join(cmd, " ")))
	return cmd
}
