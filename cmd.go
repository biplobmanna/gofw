package main

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

var mu sync.Mutex
var currentRunning *exec.Cmd

// shutdown when the child is closed
var shutdownCh = make(chan struct{})
var shutdownOnce sync.Once


func cmdRunner(meta watchMeta) {
	// mutex lock
	mu.Lock()
	defer mu.Unlock() // defer close on function end

	// print hint of function running
	log.Println("run:", meta.path, meta.cmd)


	// check if any currently running command
	if currentRunning != nil && currentRunning.Process != nil {
		// restore terminal control to gofw before killing the child
		unix.IoctlSetPointerInt(int(os.Stdin.Fd()), unix.TIOCSPGRP, syscall.Getpgrp())

		syscall.Kill(-currentRunning.Process.Pid, syscall.SIGKILL)
		currentRunning.Wait()
		currentRunning = nil
	}

	// create a commnd, and run it
	currentRunning = exec.Command("sh", "-c", meta.cmd)
	currentRunning.Dir = meta.path
	currentRunning.Stdin = os.Stdin
	currentRunning.Stdout = os.Stdout
	currentRunning.Stderr = os.Stderr

	// creating pipes for stdout, stderr
	// since by default, any new process group is started in the BG
	// the TTY will suspend the process if it tries to write to the TTY
	// stdout, _ := currentRunning.StdoutPipe()
	// stderr, _ := currentRunning.StderrPipe()

	// make shell the leader of the new process group
	currentRunning.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// start the command
	if err := currentRunning.Start(); err != nil {
		log.Printf("Failed to start command: %s\n", meta.cmd)
		os.Exit(1)
	}

	// give the child's process group terminal control so it can read from stdin
	// Process is only populated after Start()
	unix.IoctlSetPointerInt(int(os.Stdin.Fd()), unix.TIOCSPGRP, currentRunning.Process.Pid)

	// proxy the stdout, stderr to current terminal
	// go io.Copy(os.Stdout, stdout)
	// go io.Copy(os.Stderr, stderr)

	// wait for the command to complete, in the backgroud
	cmd := currentRunning // capture for local run
	go func(){
		err := cmd.Wait()

		// if the child exited natuarally, i.e., was not restarted when another event ran
		mu.Lock()
		exitedNaturally := currentRunning == cmd
		mu.Unlock()

		if exitedNaturally {
			// when the child exits naturally, restore terminal control to gofw
			unix.IoctlSetPointerInt(int(os.Stdin.Fd()), unix.TIOCSPGRP, syscall.Getpgrp())

			// only kill parent, if child was killed by a signal (CTRL+C, kill, etc...)
			// The logic:
			//  - err == nil (child exited with code 0) → keep watching
			//  - err != nil but Signaled() == false (child crashed with non-zero exit) → keep watching
			//  - Signaled() == true (child received SIGINT, SIGKILL, SIGTERM, etc.) → close parent
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					// handle shell induced exit, i.e., exitCode = 128 + N
					// where N is the signal code
					if status.Signaled()  || status.ExitStatus() >= 128 {
						shutdownOnce.Do(func(){
							close(shutdownCh)
						})
					}
				}
			}
		}
	}()
}
