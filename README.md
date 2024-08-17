## Testing

### Run Semantic Release

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" stdout`

### Print the .releaserc.json file used:

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" with-exec --args "cat,.releaserc.json" stdout`

### Export the .releaserc.json file used:

`dagger -i call with-config --file .releaserc.json release --dir . --token env:GH_TOKEN --provider "github" file --path ".releaserc.json" export --path "out.json"`


