# gofw

A file/folder watcher CLI tool written in Go. It watches a directory tree recursively for filesystem changes and re-runs a shell command automatically whenever a change is detected.

## Features

- **Recursive watching** — monitors the entire directory tree under the given path, including new subdirectories created after startup
- **PTY execution** — runs commands inside a pseudo-terminal so programs that require an interactive terminal (colour output, progress bars, curses UIs) render correctly
- **Debounced re-runs** — collapses rapid bursts of events (e.g. a formatter rewriting many files at once) into a single command invocation with a 500 ms quiet period
- **Immediate first run** — executes the command once on startup without waiting for a change
- **Clean shutdown** — kills the active child process and restores the terminal state on SIGINT/SIGTERM (Ctrl-C)
- **Terminal safety** — restores the terminal to a sane state on startup in case a previous run crashed mid-execution

## Usage

```
gofw [-p <path>] -c <command>
```

| Flag | Description | Default |
|------|-------------|---------|
| `-p` | Root path to watch | Current working directory |
| `-c` | Shell command to re-run on changes | *(required)* |

The command is executed via `sh -c`, so any shell syntax works.

### Examples

```bash
# Re-run `make build` whenever anything in the current directory changes
gofw -c "make build"

# Watch a specific directory and run tests on changes
gofw -p ./pkg -c "go test ./..."

# Watch sources and restart a dev server
gofw -p ./src -c "node server.js"
```

## Installation

```bash
go install github.com/biplobmanna/gofw@latest
```

Or build from source:

```bash
git clone https://github.com/biplobmanna/gofw
cd gofw
go build -o gofw .
```
