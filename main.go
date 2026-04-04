package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	// ignore all SIGTTOU
	signal.Ignore(syscall.SIGTTOU)

	// restore sane terminal settings in case a previous run left them broken
	if t, err := unix.IoctlGetTermios(int(os.Stdin.Fd()), unix.TIOCGETA); err == nil {
		t.Lflag |= unix.ISIG | unix.ICANON | unix.ECHO
		unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TIOCSETA, t)
	}

	// get the metadata from the commandline
	meta := argparse()

	// call the watcher
	watcher(meta)
}
