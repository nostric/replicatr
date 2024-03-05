package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Hubmakerlabs/replicatr/pkg/ic/agent"
	"mleku.dev/git/nostr/context"
	"mleku.dev/git/nostr/event"
	"mleku.dev/git/nostr/filter"
	"mleku.dev/git/nostr/keys"
	"mleku.dev/git/nostr/kind"
	"mleku.dev/git/nostr/tag"
	"mleku.dev/git/nostr/timestamp"
	"mleku.dev/git/slog"
)

func createRandomEvent(i int) (e *event.T) {
	e = &event.T{
		CreatedAt: timestamp.T(time.Now().Unix()),
		Kind:      kind.T(rand.Intn(500)),
		Tags:      []tag.T{{"tag1", "tag2"}, {"tag3"}},
		Content:   fmt.Sprintf("This is a random event content %d", i),
	}

	err := e.Sign(keys.GeneratePrivateKey())
	if err != nil {
		log.E.F("unable to create random event number %d: %v", i, err)
	}

	return
}

var log, chk = slog.New(os.Stderr)

func main() { // arg1 = canister address, arg2 = canisterID
	if len(os.Args) < 3 {
		log.E.Ln("minimum 2 args required: <canister address> <canister ID>")
		return
	}
	// Initialize the agent with the configuration for a local replica
	a, err := agent.New(os.Args[2], os.Args[1])
	if err != nil {
		log.E.F("failed to initialize agent: %v\n", err)
		return
	}
	// Create and save random events
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			log.D.Ln("creating random event")
			ev := createRandomEvent(i)
			if err = a.SaveEvent(nil, ev); chk.E(err) {
				log.E.F("Failed to save event %d: %v", i, err)
				wg.Done()
				return
			}
			log.I.F("Event %s saved successfully", ev.ID)
			wg.Done()
		}(i)
	}
	wg.Wait()
	log.I.Ln("retrieving results")
	// Create a filter to query events
	s := timestamp.Now() - 24*60*60 // for one day before now
	since := s.Ptr()
	until := timestamp.Now().Ptr()
	l := 10
	limit := &l
	f := filter.T{
		Since:  since,
		Until:  until,
		Limit:  limit,
		Search: "random",
	}
	// Query events based on the filter
	events := make(chan *event.T, 1)
	go func() {
		err = a.QueryEvents(context.Bg(), events, &f)
		if err != nil {
			log.E.Ln("Failed to query events:", err)
			return
		}
		close(events)
	}()
	log.I.Ln("receiving events")
	for ev := range events {
		log.I.F("ID: %s, Content: %s", ev.ID, ev.Content)
	}
}
