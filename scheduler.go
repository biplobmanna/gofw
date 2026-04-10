package main

import (
	"sync"
	"time"
)

// debounceTimer is the quiet period that must elapse after the last filesystem
// event before the command is actually re-run. Rapid bursts of events (e.g. a
// formatter rewriting many files at once) are collapsed into a single run.
const debounceTimer = 500 * time.Millisecond

var (
	// timerMu guards timer and triggerCounter. It is intentionally separate
	// from mu (which guards the running process) to avoid lock-order issues.
	timerMu sync.Mutex

	// triggerCounter is incremented on every call to scheduleCmdRunner. The
	// AfterFunc closure captures the counter value at scheduling time and
	// skips execution if a newer event has since been scheduled.
	triggerCounter int

	// timer is the pending debounce timer. It is reset on every new event and
	// set to nil after it fires or is stopped.
	timer *time.Timer
)

// scheduleCmdRunner debounces filesystem events. Each call cancels the
// previous pending timer and starts a new one for debounceTimer. Only the
// closure whose counter matches triggerCounter at fire time calls cmdRunner,
// ensuring intermediate events do not cause redundant runs.
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
