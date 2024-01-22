# gcyb (Go Clean Your Branches)

`gcyb` is a CLI tool for Git designed to detect and, optionally, delete branches that have already been merged into your current branch/HEAD. This tool ensures that remote branches are not deleted or updated without your explicit permission. The commands for reading and deleting branches are separated to avoid unintentional cleaning of important branches.

## Features

- Listing Merged Branches: Lists all the branches that have already been merged into the current branch in a table format.
- Deletable Branches: Identifies branches eligible for deletion based on being already merged into your branch. Stale branches, which have not received commits for an extended period, are highlighted but left for user analysis.
- Interactive Deletion: Allows you to interactively choose which branches you want to delete.
- Safety First: Prioritizes safety by not automatically deleting or updating remote branches.

## Commands

1. `gcyb`

The default execution of the CLI. It will render a table with each branch that could be deleted, and the reason of it. It does not clean or execute a delete, it only displays.

```bash
gcyb
```

2. `gcyb clean`

Deletes every branch that is considered 'deletable'. To be eligible for deletion, a branch must be already merged into your current branch.

```bash
gcyb clean
```

3. `gcyb pick`

Interactively select which branches you want to delete. The command will display a list of deletable branches, and you can choose the ones to delete.

```bash
gcyb pick
```

4. `gcyb help`

Displays information on each command or flag of gcyb.

```bash
gcyb help

||

gcyb -h

||

gcyb --help
```

## Flags

1. `repoPath`

Can specify the path of a local git repository, instead of using the working directory (current dir).

```bash
gcyb -r path/to/repo

||

gcyb --repo path/to/repo
```

## Building from Source

Ensure you have [Go](https://go.dev/) installed on your system.

```md
# Clone the repository

git clone https://github.com/lucaspiritogit/gcyb.git

# Change to the project directory

cd gcyb

# Build the binary

go build

||

go build -o gcyb

# Run the binary

./gcyb
```

## Verify checksum of binary

Downloading an .exe could and should be scary.

### Windows

To verify the integrity of the gcyb.exe binary on Windows, you can use the Get-FileHash command. Open a PowerShell window and run the following command:

```pwsh
Get-FileHash SHA256 gcyb.exe
```

### Linux

On Linux, you can use the sha256sum command to verify the checksum of the gcyb binary. Open a terminal and run the following command:

```bash
sha256sum gcyb
```

Compare the output checksum with the one provided in the [Releases](https://github.com/lucaspiritogit/gcyb/releases) page of this project.

## Contributing

If you found this repo and you want to contribute to this project, please, feel free to open issue or PR's! Feedback and contributions are welcome.
