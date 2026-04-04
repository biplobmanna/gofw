package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var mu sync.Mutex
var runningCmd *exec.Cmd
var runningPtmx *os.File
var runningOldState *term.State

var winch chan os.Signal

func winchTriggerSetup() {
	winch = make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
}

func cmdRunner(meta watchMeta) {
	// mutex lock
	mu.Lock()
	defer mu.Unlock() // defer close on function end

	// check if runningCmd, runningPtmx and close accordingly
	if runningCmd != nil && runningCmd.Process != nil {
		// kill the shell
		syscall.Kill(-runningCmd.Process.Pid, syscall.SIGKILL)
		runningCmd.Wait()

		// close ptmx if running
		if runningPtmx != nil {
			runningPtmx.Close()
		}

		// restore the term oldState
		if runningOldState != nil {
			term.Restore(int(os.Stdin.Fd()), runningOldState)
		}

		// make pointers nil
		runningCmd = nil
		runningPtmx = nil
		runningOldState = nil
	}

	// print hint of function running
	log.Println("run:", meta.path, meta.cmd)

	// Run Command
	runningCmd = exec.Command("sh", "-c", meta.cmd)
	runningCmd.Dir = meta.path

	// create a new pty, with the cmd to run
	var err error
	runningPtmx, err = pty.Start(runningCmd)
	if err != nil {
		mu.Unlock()
		log.Fatal("Couldn't start PTY")
	}

	// setup window change handler goroutine
	// only setup once
	if winch == nil {
		winchTriggerSetup()
		go func() {
			for range winch {
				// only if there's a PTMX running
				mu.Lock()
				ptmx := runningPtmx
				mu.Unlock()
				if ptmx != nil {
					if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
						log.Printf("Failed to resize PTY: %s", err)
					}
				}
			}
		}()
	}
	// trigger the WINCH signal, but non blocking
	select {
	case winch <- syscall.SIGWINCH:
	default:
	}

	// set stdin in raw mode
	runningOldState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		mu.Unlock()
		log.Fatalf("Failed to set PTY in raw mode: %s", err)
	}

	// copy stdin to pty, and pty to stdout
	go io.Copy(runningPtmx, os.Stdin)
	go io.Copy(os.Stdout, runningPtmx)

	// wait to close
	go func(cmd *exec.Cmd) {
		// exit the command; mostly exits for restarts
		cmd.Wait()

		// checks if actual exit is necessary
		mu.Lock()
		exitedNaturally := cmd == runningCmd
		mu.Unlock()

		if exitedNaturally {
			// close the runningPtmx
			mu.Lock()
			ptmx := runningPtmx
			state := runningOldState
			runningCmd = nil
			runningPtmx = nil
			runningOldState = nil
			mu.Unlock()

			// restore terminal state, close ptmx
			if ptmx != nil {
				ptmx.Close()
			}
			if state != nil {
				term.Restore(int(os.Stdin.Fd()), state)
			}

			// log message
			log.Printf("Child terminated gracefully...\nPress CTRL+C to terminate parent...\n")
		}
	}(runningCmd)
}
