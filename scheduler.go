package main

import (
	"sync"
	"time"
)

const debounceTimer = 500 * time.Millisecond

var (
	timerMu        sync.Mutex
	triggerCounter int
	timer          *time.Timer
)

func scheduleCmdRunner(meta watchMeta) {
	// log.Println("scheduleCmdRunner > ")

	// lock the mutex
	timerMu.Lock()
	defer timerMu.Unlock() // unlock on function close

	// close the previous timer
	if timer != nil {
		timer.Stop()
		timer = nil
	}

	// update the counter
	triggerCounter++
	timerCounter := triggerCounter

	// call the afterFunc
	timer = time.AfterFunc(debounceTimer, func() {
		timerMu.Lock()
		defer timerMu.Unlock()

		// only latest function will run
		if timerCounter == triggerCounter {
			cmdRunner(meta)
		}
	})
}
