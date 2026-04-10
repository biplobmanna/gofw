//go:build darwin

package main

import "golang.org/x/sys/unix"

// restoreSaneTerminal sets ISIG, ICANON, and ECHO on the terminal identified
// by fd. It is called at startup to recover from a previous run that may have
// left the terminal in raw mode after a crash. On Darwin the termios structure
// is retrieved and applied via TIOCGETA/TIOCSETA.
func restoreSaneTerminal(fd int) {
	if t, err := unix.IoctlGetTermios(fd, unix.TIOCGETA); err == nil {
		t.Lflag |= unix.ISIG | unix.ICANON | unix.ECHO
		unix.IoctlSetTermios(fd, unix.TIOCSETA, t)
	}
}
