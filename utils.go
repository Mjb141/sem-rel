package main

import (
	"context"
	"dagger/sem-rel/internal/dagger"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

func (m *SemRel) AddBranchToReleaseRc(ctx context.Context, dir *dagger.Directory, branches []Branch) ([]Branch, error) {
	currentBranch, err := dag.GitInfo(dir).Branch(ctx)
	if err != nil {
		log.
			Error().
			Err(err).
			Msg("Could not get the current branch name")

		return branches, err
	}

	var branchNamesInReleaseRc []string
	for _, branch := range branches {
		log.
			Info().
			Str("Branch", branch.Name).
			Msg("Branch found")

		branchNamesInReleaseRc = append(branchNamesInReleaseRc, branch.Name)
	}

	currentBranchInReleaseBranches := slices.Contains(branchNamesInReleaseRc, currentBranch)

	if currentBranchInReleaseBranches {
		log.
			Info().
			Msg(fmt.Sprintf("Branch %s already found in config.Branches", currentBranch))

		return branches, nil
	}

	log.
		Info().
		Msg(fmt.Sprintf("Adding branch %s to config.Branches", currentBranch))

	return append(branches, Branch{currentBranch, "", false}), nil
}

func (m *SemRel) RemoveGitPlugin(ctx context.Context, dir *dagger.Directory, plugins []Plugin) []Plugin {
	for pluginIndex, plugin := range plugins {
		pluginType := reflect.TypeOf(plugin)
		switch pluginType.Kind() {
		case reflect.String:
			log.
				Debug().
				Str("string", fmt.Sprintf("%s", plugin)).
				Msg("Plugins")

			if plugin.(string) == "@semantic-release/github" || plugin.(string) == "@semantic-release/gitlab" {
				log.
					Debug().
					Msg(fmt.Sprintf("Removing pluginIndex %d (%s) from plugins", pluginIndex, plugin))

				plugins = slices.Delete(plugins, pluginIndex, pluginIndex+1)

				log.
					Debug().
					Str("array", fmt.Sprintf("%s", plugins)).
					Msg("Plugins")

				return plugins
			}

		case reflect.Array, reflect.Slice:
			log.
				Debug().
				Str("array", fmt.Sprintf("%s", plugin)).
				Msg("Plugins")

		default:
			log.
				Debug().
				Str("other", fmt.Sprintf("%s", plugin)).
				Msg("Plugins")
		}
	}

	return plugins
}

func (m *SemRel) SemanticReleaseCommand(ctx context.Context, dryRun, checkIfCi bool) []string {
	cmd := []string{"semantic-release"}

	if dryRun {
		log.
			Debug().
			Str("option", "--dry-run").
			Msg("Added option to 'semantic-release' command")

		cmd = append(cmd, "--dry-run")
	}

	if !checkIfCi {
		log.
			Debug().
			Str("option", "--no-ci").
			Msg("Added option to 'semantic-release' command")

		cmd = append(cmd, "--no-ci")
	}

	log.
		Debug().
		Msg(fmt.Sprintf("Returning 'semantic-release' command: '%s'", strings.Join(cmd, " ")))

	return cmd
}
