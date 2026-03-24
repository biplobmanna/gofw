package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// go routine to watch
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Remove) {
					log.Printf("File Modified: %v", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Fatal(err)
			}
		}
	}()

	// path to watch
	path := "."
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watching for changes in: %v", path)

	// block the main gorouting forever
	<-make(chan struct{})

}
