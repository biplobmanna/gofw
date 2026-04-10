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

// mu guards the three running-process globals below. It must not be held by
// any caller of cmdRunner, because cmdRunner acquires mu itself at the top.
var mu sync.Mutex

// runningCmd is the currently executing child process, or nil if none is running.
var runningCmd *exec.Cmd

// runningPtmx is the PTY master file attached to runningCmd, or nil.
var runningPtmx *os.File

// runningOldState is the saved terminal state from before stdin was put into
// raw mode for the current child. It is restored when the child exits or is killed.
var runningOldState *term.State

// winch is the channel that receives SIGWINCH signals so the PTY can be
// resized to match the terminal window. It is initialised once, lazily, the
// first time cmdRunner is called.
var winch chan os.Signal

// winchTriggerSetup registers the process to receive SIGWINCH on the winch
// channel. It must be called exactly once before the SIGWINCH forwarding
// goroutine is started.
func winchTriggerSetup() {
	winch = make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
}

// cmdRunner kills any currently running child process, then starts meta.cmd
// via "sh -c" inside a PTY so that programs requiring an interactive terminal
// work correctly. It sets stdin to raw mode for the duration of the child's
// life, pipes stdin/stdout through the PTY, and spawns a goroutine that
// restores the terminal when the child exits naturally.
//
// cmdRunner acquires mu at the top and must never be called while mu is
// already held by the same goroutine.
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
