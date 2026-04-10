package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/term"
)

// watchMeta carries the configuration supplied by the user via CLI flags.
type watchMeta struct {
	// path is the root directory to watch recursively.
	path string
	// cmd is the shell command re-run on every detected change.
	cmd string
}

// watcher is the main run loop. It defaults path to the current working
// directory when none is provided, registers the entire directory tree with
// fsnotify, runs the command once on startup, and then blocks until SIGINT or
// SIGTERM is received. On shutdown it kills the active child process and
// restores the terminal state.
func watcher(meta watchMeta) {
	// path to watch; if no path is provided
	// use the current working directory
	if meta.path == "" {
		var err error
		meta.path, err = os.Getwd()
		if err != nil {
			log.Fatal("Failed to fetch the current working directory!")
		}
	}

	// new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	// defer close as soon as the watcher() returns
	defer watcher.Close()

	// go routine to watch
	go watching(meta, watcher)

	// run the command for the first time
	cmdRunner(meta)

	// use the created watcher for the first time
	err = watchRecursive(meta.path, watcher)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watching for changes in: %s\n", meta.path)

	// block the main gorouting forever
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// exit on SIGTERM
	<-sig
	mu.Lock()
	if runningCmd != nil && runningCmd.Process != nil {
		syscall.Kill(-runningCmd.Process.Pid, syscall.SIGKILL)
		runningCmd.Wait()
		if runningPtmx != nil {
			runningPtmx.Close()
		}
		if runningOldState != nil {
			term.Restore(int(os.Stdin.Fd()), runningOldState)
		}
	}
	mu.Unlock()
	log.Println("shutting down gracefully...")
}

// watching is the event loop that consumes fsnotify events in a goroutine.
// Write and Remove events trigger a debounced command re-run immediately.
// Create events additionally register any new subdirectory with the watcher
// before scheduling a re-run, ensuring the watch tree stays up-to-date as
// directories are added.
func watching(meta watchMeta, watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				// schedule the command runner
				scheduleCmdRunner(meta)
			} else if event.Has(fsnotify.Create) {
				info, err := os.Stat(event.Name)
				if err != nil {
					log.Fatal(err)
				} else if info.IsDir() {
					watchRecursive(event.Name, watcher)
				}
				// schedule the command runner
				scheduleCmdRunner(meta)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Fatal(err)
		}
	}
}

// watchRecursive walks the directory tree rooted at path and registers every
// subdirectory with watcher. Files are ignored because fsnotify watches at the
// directory level and reports events for all files within a watched directory.
func watchRecursive(path string, watcher *fsnotify.Watcher) error {
	return filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(newPath)
		}
		return nil
	})
}
