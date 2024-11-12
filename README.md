## Testing

### Run Semantic Release with defaults:

`dagger call release --token env:GH_TOKEN stdout`

### Run Semantic Release on a non-configured branch:

**Note:** Branch must exist on remote source. Local-only branches will return "This test run was triggered on the branch <name>, while semantic-release is configured to only publish from main."

`dagger call configure --add-current-branch release --token env:GH_TOKEN stdout`

### Print the .releaserc.json file used (uses `jq` to pretty print JSON):

`dagger call configure --add-current-branch release --token env:GH_TOKEN file --path ".releaserc.json" contents | jq`

### Export the .releaserc.json file used:

`dagger call configure --add-current-branch release --token env:GH_TOKEN file --path ".releaserc.json" export --path ".releaserc.json.modified"`

`cat .releaserc.json.modified | jq`
