//go:build darwin

package main

import "golang.org/x/sys/unix"

func restoreSaneTerminal(fd int) {
	if t, err := unix.IoctlGetTermios(fd, unix.TIOCGETA); err == nil {
		t.Lflag |= unix.ISIG | unix.ICANON | unix.ECHO
		unix.IoctlSetTermios(fd, unix.TIOCSETA, t)
	}
}
