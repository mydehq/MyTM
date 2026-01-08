<div align="center">
<img src="./app/icons/icon.png" alt="MyTM" width="80">
<h1>MyTM</h1>

MyTM or My Theme Manager is a [MyCTL](https://github.com/mydehq/myctl) plugin for desktop theme management.
<br>
Supports gtk, qt, kde, rofi & others through templates

</div>

---

> [!CAUTION]  
> **Experimental Branch**  
> This is rewrite of the bash version for testing performence/feasability.  
> The stable Bash version resides in the [main](../tree/main) branch.  

## ⚡ Performance

The Go port significantly outperforms the Bash script, especially as the number of themes grows.

| Scenario           | Load        | Bash Time | Go Time | Speedup |
| :----------------- | :---------- | :-------- | :------ | :------ |
| **Realistic Load** | 12 Themes   | ~1.80s    | ~0.03s  | **54x** |
| **Mid Load**       | 100 Copies  | ~13.40s   | ~0.17s  | **78x** |
| **High Load**      | 500 Copies  | ~69.00s   | ~0.80s  | **86x** |
| **Stress Test**    | 1000 Copies | ~134.00s  | ~1.50s  | **87x** |

This massive scalability difference is because Go handles file I/O, archiving, and compression in-memory within a single process, avoiding the overhead of spawning thousands of subprocesses (`tar`, `jq`, `yq`, `cp`, `mkdir`) that the Bash script requires for every single file operation.

### 📦 Efficiency

We also measured the output size to ensure no bloat:

- **Bash Output**: Identical to Bash (~8-9MB for 500-600 themes).
- **Result**: **100% size parity.** The Go port produces optimized archives just like the original.



## 🚀 Getting Started

### Prerequisites

- [Go 1.21+](https://go.dev/dl/) installed.

### Installation

You can build the binary from source:

```bash
# Clone the repository
git clone https://github.com/mydehq/mytm.git
cd mytm/app

# Build the binary
go build -o ../mytm main.go
cd ..
```

Now you have an executable `mytm` in your current directory.

### Usage

#### Build Themes

Builds all themes in the input directory and updates the repository index.

```bash
# Verify usage
./mytm build --help

# Run build
./mytm build -c config.yml -i MyTM/themes -o dist
```

#### Validate Themes

Checks if the theme directories validity without packaging them.

```bash
./mytm validate -i MyTM/themes
```

---

### 📚 Some information

If you are using this project to learn Go, I recommend reading the files in this specific order to understand how the application flows:

1.  **[main.go](./app/main.go)**: The entry point. Learn how `package main` works and how we hand off control to the CLI library.
2.  **[cmd/root.go](./app/cmd/root.go)**: See how **Cobra** (the CLI library) is initialized and where flags are defined.
3.  **[cmd/build.go](./app/cmd/build.go)**: The "Glue" code. This file orchestrates the entire build process. Read the `Run` function step-by-step.
4.  **[internal/config/config.go](./app/internal/config/config.go)**: Learn how struct tags (`yaml:"..."`) map configuration files to Go structs.
5.  **[internal/utils/utils.go](./app/internal/utils/utils.go)**: Dive into `CreateTarGz` to see real-world File I/O, recursing directories, and handling specific Tar headers.
6.  **[internal/theme/theme.go](./app/internal/theme/theme.go)**: See how we validate inputs using RegEx and robust error handling.
7.  **[internal/repo/repo.go](./app/internal/repo/repo.go)**: Deep dive into JSON manipulation—reading a file, unmarshalling to a struct, modifying it (appending/truncating slices), and writing it back.

### Project Structure

This project follows a standard Go CLI project layout (Standard Go Project Layout).

### Directory Benchmark

| Directory          | Purpose                                                                                                                       |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| `cmd/`             | Contains the commands for the CLI. Each file here usually corresponds to a subcommand (e.g., `cmd/build.go` -> `mytm build`). |
| `internal/`        | Private application code. These libraries cannot be imported by other projects.                                               |
| `internal/config/` | Handles loading and parsing `config.yml` using `gopkg.in/yaml.v3`.                                                            |
| `internal/theme/`  | Logic for validating theme structure and parsing `theme.yml`.                                                                 |
| `internal/utils/`  | General helper functions (Hashing, File I/O, Tar Gzipping).                                                                   |
| `internal/repo/`   | Logic for managing `index.json` and `versions.json`.                                                                          |
| `main.go`          | The entry point. It simply calls `cmd.Execute()`.                                                                             |

### Key Libraries Used

- **[Cobra](https://github.com/spf13/cobra)**: The industry standard framework for building CLI applications in Go. It handles flags, subcommands, and help generation.
- **[YAML](https://gopkg.in/yaml.v3)**: Used for robust YAML parsing.

### Interesting Code Points

- **Recursive Directory Walking**: Check `internal/utils/utils.go` (`CreateTarGz`) to see how we walk a directory tree to create an archive.
- **Symlink Handling**: Also in `CreateTarGz`, notice how we explicitly check for symlinks to preserve them in the archive.
- **Dependency Injection**: Notice how `build.go` orchestrates the other internal packages (`config`, `repo`, `utils`) to complete the task.

## 🤝 Parity with Bash Script

This Go implementation has been verified to produce **bit-exact** output parity with the original Bash script (excluding timestamps) for:

- `.tar.gz` structure and content.
- `index.json` structure.
- `index.html` and `README.md` generation.

### 🔍 Why is Go so much faster?

The Bash script suffers from **process overhead**. For _every single theme_, it spawns multiple new processes (programs):

1.  `yq`: To read the version. Even though `yq` is fast, the OS has to **spawn a new process** and initialize the Go runtime for it 600 separate times.
2.  `tar`: To create the archive.
3.  `sha256sum`: To generate the hash.
4.  `jq`: Multiple times to read/update JSON files.

For 600 themes, that's over **2,500 process launches**. Operating systems take time to start each program.

**Go does all of this in a single running process.** It parses YAML/JSON, creates TarGz archives, and calculates Hashes directly in memory using standard libraries, eliminating 99% of the overhead.
