package http

import (
	"context"
	"log"
	"sync"

	"ts6-viewer/internal/config"
	"ts6-viewer/internal/ts6"
)

// eventQueueSize bounds how many real-time events may be queued while the
// worker is folding one observation. Extra events are coalesced (dropped):
// the in-flight fetch already reflects the current server state, and the
// 30s sampler plus the client-database safety net guarantee the visit is
// recorded regardless.
const eventQueueSize = 16

var (
	eventMu     sync.Mutex // guards eventCancel, eventDone, eventQueue
	eventCancel context.CancelFunc
	eventDone   chan struct{}
	eventQueue  chan string // event names waiting for a statistics fold
)

// StartEventStatistics wires real-time ServerQuery event capture into the
// statistics pipeline: a client enter/leave is folded into the state
// immediately instead of waiting for the next 30s sampler tick. It is
// idempotent: starting twice is a no-op.
//
// If the SSH connection cannot be established (server down, bad
// credentials) it logs and returns without spawning anything; the 30s
// sampler remains the capture path and will pick the connection up on its
// next tick.
func StartEventStatistics(cfg *config.Config) {
	eventMu.Lock()
	defer eventMu.Unlock()

	if eventCancel != nil {
		// Already running.
		return
	}

	sshClient, err := ts6.GetPersistentClient(cfg, cfg.Teamspeak6.ServerID)
	if err != nil {
		log.Printf("[Events] SSH connection unavailable, real-time event capture disabled (30s sampler remains the capture path): %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	eventCancel = cancel
	eventDone = done
	eventQueue = make(chan string, eventQueueSize)

	// The callback runs in the SSH reader goroutine, so it must be fast and
	// non-blocking: enqueue the event name and let the worker do the fetch.
	// SetEventCallback also sends "servernotifyregister event=server" once
	// (and re-sends it automatically after every reconnect).
	sshClient.SetEventCallback(func(ev ts6.ServerEvent) {
		select {
		case eventQueue <- ev.Name:
		default:
			log.Println("[Events] Queue full, coalescing event (worker is already catching up)")
		}
	})

	log.Println("[Events] Real-time event capture active")
	go runEventWorker(cfg, ctx, done)
}

// StopEventStatistics cancels the event worker and waits for it to exit.
// Idempotent: calling it without a start, or twice, is a no-op.
func StopEventStatistics() {
	eventMu.Lock()
	cancel := eventCancel
	done := eventDone
	if cancel != nil {
		eventCancel = nil
		eventDone = nil
		eventQueue = nil
	}
	eventMu.Unlock()

	if cancel == nil {
		return // never started
	}

	cancel()
	<-done
	log.Println("[Events] Event worker stopped")
}

// runEventWorker folds queued real-time events into the statistics state.
// It is deliberately conservative about load: after one fetch it drains
// (coalesces) the rest of the queue, because that fetch already reflects
// the current online state and the 30s sampler will catch anything the
// coalescing dropped.
func runEventWorker(cfg *config.Config, ctx context.Context, done chan struct{}) {
	defer close(done)

	for {
		select {
		case <-ctx.Done():
			return
		case name := <-eventQueue:
			log.Printf("[Events] %s: folding statistics now\n", name)
			recordOne(cfg, "event:"+name)

			// Coalesce: drain events queued while the fetch was in flight.
			// The fetch above already captured the latest state, so each
			// drained event is redundant (the client database safety net
			// covers any edge case).
			coalesced := 0
		drain:
			for {
				select {
				case <-eventQueue:
					coalesced++
				default:
					break drain
				}
			}
			if coalesced > 0 {
				log.Printf("[Events] Coalesced %d queued event(s)\n", coalesced)
			}
		}
	}
}
