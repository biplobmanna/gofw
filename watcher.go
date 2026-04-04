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

type watchMeta struct {
	path string
	cmd  string
}

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
