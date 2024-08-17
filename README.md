## Testing

### Run Semantic Release:

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" stdout`

### Run Semantic Release on a non-configured branch:

**Note:** Branch must exist on remote source. Local-only branches will return "This test run was triggered on the branch <name>, while semantic-release is configured to only publish from main."

`dagger -i call with-config --file .releaserc.json --branch "test" release --dir . --token env:GH_TOKEN --provider "github" stdout`

### Print the .releaserc.json file used:

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" with-exec --args "cat,.releaserc.json" stdout`

### Export the .releaserc.json file used:

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" file --path ".releaserc.json" export --path "out.json"`


