package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func watcher(path string) {
	// path to watch; if no path is provided
	// use the current working directory
	if path == "" {
		var err error
		path, err = os.Getwd()
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
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
					log.Printf("%v: %v", event.Op.String(), event.Name) // replaced with running the cmd
				} else if event.Has(fsnotify.Create) {
					log.Printf("%v: %v", event.Op.String(), event.Name) // replaced with running the cmd
					info, err := os.Stat(event.Name)
					if err != nil {
						log.Fatal(err)
					} else if info.IsDir() {
						watchRecursive(event.Name, watcher)
					}

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Fatal(err)
			}
		}
	}()

	// build the watcher
	err = watchRecursive(path, watcher)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watching for changes in: %v", path)

	// block the main gorouting forever
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("shutting down gracefully...")

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
