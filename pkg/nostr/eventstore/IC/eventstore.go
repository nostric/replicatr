package IC

import (
	"os"

	"github.com/Hubmakerlabs/replicatr/pkg/ic/agent"
	"mleku.dev/git/nostr/context"
	"mleku.dev/git/nostr/event"
	"mleku.dev/git/nostr/eventstore/badger"
	"mleku.dev/git/nostr/filter"
	"mleku.dev/git/nostr/kinds"
	"mleku.dev/git/nostr/relayinfo"
	"mleku.dev/git/slog"
)

var log, chk = slog.New(os.Stderr)

type Backend struct {
	// Badger backend must be populated
	Badger *badger.Backend
	IC     *agent.Backend
}

// Init sets up the badger event store and connects to the configured IC
// canister.
func (b *Backend) Init(inf *relayinfo.T, params ...string) (err error) {
	if len(params) < 2 {
		return log.E.Err("not enough parameters for eventstore Init, got %d, "+
			"require %d: %v", len(params), 2, params)
	}
	canisterId, addr := params[0], params[1]
	if err = b.Badger.Init(inf); chk.D(err) {
		return
	}
	if b.IC, err = agent.New(canisterId, addr); chk.E(err) {
		return
	}
	return
}

// Close the connection to the database.
func (b *Backend) Close() { b.Badger.Close() }

// Serial returns the serial code for the database.
func (b *Backend) Serial() []byte { return b.Badger.Serial() }

// CountEvents returns the number of events found matching the filter.
func (b *Backend) CountEvents(c context.T, f *filter.T) (count int, err error) {
	var forBadger, forIC kinds.T
	for i := range f.Kinds {
		if kinds.IsPrivileged(f.Kinds[i]) {
			forBadger = append(forBadger, f.Kinds[i])
		} else {
			forIC = append(forIC, f.Kinds[i])
		}
	}
	icFilter := f.Duplicate()
	f.Kinds = forBadger
	icFilter.Kinds = forIC
	if count, err = b.Badger.CountEvents(c, f); chk.E(err) {
		return
	}
	// todo: this will be changed to call the IC count events implementation
	return b.Badger.CountEvents(c, icFilter)
}

// DeleteEvent removes an event from the event store.
func (b *Backend) DeleteEvent(c context.T, evt *event.T) (err error) {
	if kinds.IsPrivileged(evt.Kind) {
		return b.Badger.DeleteEvent(c, evt)
	}
	// todo: this will be the IC store
	return b.Badger.DeleteEvent(c, evt)
}

// QueryEvents searches for events that match a filter and returns them
// asynchronously over a provided channel.
func (b *Backend) QueryEvents(c context.T, C chan *event.T,
	f *filter.T) (err error) {

	var forBadger, forIC kinds.T
	for i := range f.Kinds {
		if kinds.IsPrivileged(f.Kinds[i]) {
			forBadger = append(forBadger, f.Kinds[i])
		} else {
			forIC = append(forIC, f.Kinds[i])
		}
	}
	icFilter := f.Duplicate()
	f.Kinds = forBadger
	icFilter.Kinds = forIC
	if len(forBadger) > 0 {
		log.I.Ln("querying relay store with filter", f.ToObject().String())
		if err = b.Badger.QueryEvents(c, C, icFilter); chk.E(err) {
		}
	}
	if len(forIC) > 0 {
		log.I.Ln("querying IC with filter", icFilter.ToObject().String())
		if err = b.IC.QueryEvents(c, C, f); chk.E(err) {
		}
	}
	return
}

// SaveEvent writes an event to the event store.
func (b *Backend) SaveEvent(c context.T, ev *event.T) (err error) {
	if kinds.IsPrivileged(ev.Kind) {
		log.I.Ln("saving event to relay store", ev.ToObject().String())
		return b.Badger.SaveEvent(c, ev)
	}
	log.I.Ln("saving event to IC", ev.ToObject().String())
	return b.IC.SaveEvent(c, ev)
}
