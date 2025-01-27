## Testing

Defaults:
- Local execution is prevented
- Dry run is set

This prevents any accidental releases.

### Dry run a Semantic Release locally:

`dagger call configure --allow-local release --token env:GH_TOKEN stdout`

### Run a Semantic Release locally:

:warning: This will trigger a release from your local machine :warning:

`dagger call configure --allow-local --dry-run=false release --token env:GH_TOKEN stdout`

### Run Semantic Release on a non-configured branch:

**Note:** Branch must exist on remote source. Local-only branches will return "This test run was triggered on the branch <name>, while semantic-release is configured to only publish from main."

`dagger call configure --add-current-branch release --token env:GH_TOKEN stdout`

### Print the .releaserc.json file used (uses `jq` to pretty print JSON):

`dagger call configure --add-current-branch release --token env:GH_TOKEN file --path ".releaserc.json" contents | jq`

### Export the .releaserc.json file used:

`dagger call configure --add-current-branch release --token env:GH_TOKEN file --path ".releaserc.json" export --path ".releaserc.json.modified"`

`cat .releaserc.json.modified | jq`

