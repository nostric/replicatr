package icp

import (
	"io"
	"sort"
	"strings"

	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/filter"
	"realy.lol/store"

	"github.com/Hubmakerlabs/replicatr/canister/candid"
)

var _ store.I = &T{}

// Init is a no op because this store is a network service
func (b *T) Init(path st) (err er) {
	return
}

// Path is a no op because this store is a network service
func (b *T) Path() (s st) {
	return
}

// Close is a no op because this store is a network service
func (b *T) Close() (err er) {
	return
}

// Nuke deletes all of the events in the store.
func (b *T) Nuke() (err er) {
	if err = b.ClearEvents(); chk.E(err) {
		return
	}
	return
}

// Import is a no op because this store is a network service
func (b *T) Import(r io.Reader) {
	return
}

// Export is a no op because this store is a network service
func (b *T) Export(c cx, w io.Writer, pubkeys ...by) {
	return
}

// Sync is a no op because this store is a network service
func (b *T) Sync() (err er) {
	return
}

func (b *T) QueryEvents(c cx, f *filter.T) (evs event.Ts, err er) {
	if f == nil {
		err = log.E.Err("nil filter for query")
		return
	}
	var candidEvents []candid.Event
	if candidEvents, err = b.GetCandidEvent(FilterToCandid(f)); err != nil {
		split := strings.Split(err.Error(), "Error: ")
		if len(split) == 2 && split[1] != "No events found" {
			log.E.F("IC error: %s", split[1])
		}
		return
	}
	log.I.Ln("got", len(candidEvents), "events")
	evs = make(event.Ts, len(candidEvents))
	for i, e := range candidEvents {
		evs[i] = CandidToEvent(&e)
	}
	// sort so the newest are first
	sort.Sort(evs)
	return
}

func (b *T) SaveEvent(c cx, ev *event.T) (err er) {
	if ev == nil {
		err = errorf.E("can't save nil event")
		return
	}
	if err = b.SaveCandidEvent(EventToCandid(ev)); chk.E(err) {
		return
	}
	return
}

func (b *T) DeleteEvent(c cx, eid *eventid.T) (err er) {
	if err = b.DeleteCandidEvent(eid.String()); chk.E(err) {
		return
	}
	return
}

func (b *T) CountEvents(c cx, f *filter.T) (count no, approx bo, err er) {
	if f == nil {
		err = log.E.Err("nil filter for count query")
		return
	}
	count, err = b.CountCandidEvent(FilterToCandid(f))
	return
}
