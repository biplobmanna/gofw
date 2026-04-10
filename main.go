// gofw watches a directory tree for filesystem changes and re-runs a shell
// command whenever a change is detected. It is designed for development
// workflows where a build or test command should re-execute automatically on
// every save.
//
// Usage:
//
//	gofw [-p <path>] [-x <command>] [-c <config>]
//
// Flags:
//
//	-p  Root path to watch (default: current working directory).
//	-x  Shell command to re-run on changes. Executed via "sh -c".
//	-c  Path to a YAML config file providing path and/or command.
//
// CLI flags take precedence: values from -p and -x override any corresponding
// fields in the config file only when the config field is empty.
//
// gofw runs the command immediately on startup, then re-runs it (debounced by
// 500 ms) whenever a Write, Remove, or Create event is detected anywhere under
// the watched tree. New subdirectories are watched automatically as they appear.
// The command runs inside a PTY so programs that require an interactive terminal
// (colour output, progress bars, etc.) behave correctly.
//
// Terminate gofw with SIGINT or SIGTERM (Ctrl-C). The currently running child
// process is killed and the terminal is restored before exit.
package main

import (
	"os"
	"os/signal"
	"syscall"
)

// main is the program entry point. It suppresses SIGTTOU (which would otherwise
// pause the process when it tries to write to the terminal from the background),
// restores the terminal to a sane cooked state in case a previous invocation
// crashed mid-run, parses CLI flags, and starts the watcher loop.
func main() {
	// ignore all SIGTTOU
	signal.Ignore(syscall.SIGTTOU)

	restoreSaneTerminal(int(os.Stdin.Fd()))

	// get the metadata from the commandline
	meta := argparse()
	// if config file is passed, then use that config file to update the meta
	parseYamlFile(&meta)

	// call the watcher
	watcher(meta)
}
