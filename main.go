package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// ignore all SIGTTOU
	signal.Ignore(syscall.SIGTTOU)

	restoreSaneTerminal(int(os.Stdin.Fd()))

	// get the metadata from the commandline
	meta := argparse()

	// call the watcher
	watcher(meta)
}
