//go:build linux

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

// restoreSaneTerminal sets ISIG, ICANON, and ECHO on the terminal identified
// by fd. It is called at startup to recover from a previous run that may have
// left the terminal in raw mode after a crash. On Linux the termios structure
// is retrieved and applied via TCGETS/TCSETS.
func restoreSaneTerminal(fd int) {
	if t, err := unix.IoctlGetTermios(fd, unix.TCGETS); err == nil {
		t.Lflag |= unix.ISIG | unix.ICANON | unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, t); err != nil {
			log.Println("Failed to reset the terminal to sane defaults...")
		}
	}
}
