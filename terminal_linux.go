//go:build linux

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func restoreSaneTerminal(fd int) {
	if t, err := unix.IoctlGetTermios(fd, unix.TCGETS); err == nil {
		t.Lflag |= unix.ISIG | unix.ICANON | unix.ECHO
		if err := unix.IoctlSetTermios(fd, unix.TCSETS, t); err != nil {
			log.Println("Failed to reset the terminal to sane defaults...")
		}
	}
}
