# Dirtree

Dirtree is a command-line utility that generates visual and machine-readable representations of directory trees. Point it at a folder and it prints a clean, configurable tree to your terminal, and can optionally export it as `.txt` or `.json`.

```
+--- main.go
+--- docs/
|   `--- readme.md
`--- src/
   +--- app.go
   `--- utils/
      `--- helper.go
```

## Features

- Terminal, text file (`tree.txt`), and JSON (`tree.json`) output, any combination at once
- Configurable connector style, spacing/density, and sort order (alphabetic, files-first/directories-first)
- Hidden file inclusion/exclusion
- Depth and element limits to keep large trees readable, with a summary line for anything truncated (e.g. `and 12 files not shown`)
- Persistent config file so you don't have to repeat flags every run
- Adjustable CPU usage via a `powerLevel` setting for large scans

## Installation

Requires Go 1.25+.

```bash
git clone <this-repo>
cd dirtree
go build
```

Or use the provided build script, which formats the code, builds the binary, and copies it to `~/.local/bin`:

```bash
./build.sh
```

Make sure `~/.local/bin` is on your `PATH`.

## Usage

```bash
dirtree [path] [flags]
```

If no path is given, dirtree uses the current working directory.

```bash
dirtree                      # tree of the current directory
dirtree ~/projects/myapp     # tree of a specific directory
dirtree --json --txt myapp   # also write tree.json and tree.txt
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--terminal` | `true` | Print the tree to the terminal |
| `--txt` | `false` | Write the tree to `tree.txt` |
| `--json` | `false` | Write the tree to `tree.json` |
| `o`, `--output` | *(input path)* | Output directory for `tree.txt` / `tree.json` |
| `--density` | `3` | Tree spacing; `1` is dense, `5` is spacious |
| `t`, `--treeset` | `2` | Connector style; see [Connector Sets](#connector-sets) below |
| `--filesfirst` | `true` | List files before directories |
| `--alphabetic` | `true` | Sort entries alphabetically |
| `--hidden` | `false` | Include hidden files/directories |
| `--maxdepth` | `10` | Maximum nesting depth to descend into |
| `--maxelements` | `20` | Maximum entries shown per directory before summarizing the rest |
| `--dirhints` | `true` | Include the full path in truncation summaries |
| `--powerLevel` | `m` | CPU usage: `l` (2 cores), `m` (half of available), `a` (all) |
| `-u`, `--unrestricted` | `false` | Ignore depth/element limits and use all cores (prompts for confirmation; output can get very large) |
| `-v`, `--verbose` | `false` | Enable debug logging |
| `-s`, `--saveconf` | `false` | Save the current flag values as the new default configuration |
| `--displayconf` | `false` | Print the current configuration and exit |

`--annotate` (short `-a`) and `--annotationspadding` are reserved for an upcoming annotation feature and currently have no effect.

### Connector Sets

- `1` — `└ ├ │ ─`
- `2` — `` + ` | - ``
- `3` — `| | _ |`
- `4` — `o o . . o`
- `5` — `║ ╚ ═ ║`
- `6` — `╿ ┖ ╼ ╿`
- `7` — `│ ╰ ─ │`

## Configuration

On first run, dirtree writes a default config file to:

- `$XDG_CONFIG_HOME/dirtree/config.yaml`, or
- `~/.config/dirtree/config.yaml` if `XDG_CONFIG_HOME` is unset

Any flag you pass overrides the config for that run. Pass `-s`/`--saveconf` alongside your flags to persist those values as the new defaults, e.g.:

```bash
dirtree --density 1 --connectorset 1 --maxelements 50 -s
```

Run `dirtree --displayconf` to see the currently loaded configuration.

## Development

Run the test suite with:

```bash
go test ./...
```

## Credits

Thanks a bunch to [pipeseroni](https://github.com/pipeseroni) for the [pipes.sh](https://github.com/pipeseroni/pipes.sh) project that gave me the idea for the sets 4 through 7 and for the treeset (shorthand t) flag name.

## License

Dirtree is licensed under the [GNU General Public License v3.0 or later](LICENSE).
