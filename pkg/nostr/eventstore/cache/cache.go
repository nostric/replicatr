package cache

import (
	"bytes"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/Hubmakerlabs/replicatr/pkg/nostr/context"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/event"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/eventid"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/timestamp"
	"mleku.dev/git/slog"
)

var log, chk = slog.New(os.Stderr)

type Event struct {
	JSON         []byte
	lastAccessed timestamp.T
}

type events map[eventid.T]*Event

type Encoder struct {
	mx        sync.Mutex
	events    events
	pool      *sync.Pool
	sizeLimit int
}

type access struct {
	eventid.T
	lastAccessed timestamp.T
	size         int
}

// accesses is a sort.Interface that sorts in descending order of lastAccessed
type accesses []access

func (s accesses) Len() int           { return len(s) }
func (s accesses) Less(i, j int) bool { return s[i].lastAccessed > s[j].lastAccessed }
func (s accesses) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// NewEncoder creates a cache.Encoder that maintains a cache of already decoded
// JSON bytes of events to prevent doubling up of both decoding work and JSON
// bytes usage for multiple concurrent requests of overlapping events by
// clients.
func NewEncoder(c context.T, sizeLimit int, gcTimer time.Duration) *Encoder {
	d := &Encoder{
		events:    make(events),
		sizeLimit: sizeLimit,
		pool: &sync.Pool{New: func() any {
			return make([]byte, 0, sizeLimit)
		}},
	}
	go func() {
		log.I.Ln("starting decoder cache garbage collector")
		tick := time.NewTicker(gcTimer)
		for {
			select {
			case <-c.Done():
				log.I.Ln("terminating decoder cache garbage collector")
				return
			case <-tick.C:
				var total int
				for i := range d.events {
					total += len(d.events[i].JSON)
				}
				if total > d.sizeLimit {
					// create list of cache by access time
					var accessed accesses
					for id := range d.events {
						accessed = append(accessed,
							access{
								T:            id,
								lastAccessed: d.events[id].lastAccessed,
								size:         len(d.events[id].JSON),
							})
					}
					sort.Sort(accessed)
					var last, size int
					// count off the items in descending timestamp order until the size exceeds the
					// sizeLimit.
					for ; size < d.sizeLimit; last++ {
						size += accessed[last].size
					}
					log.I.F("pruning out %d bytes of %d of cached decoded events")
					for ; last < len(accessed); last++ {
						// free the buffers so they go back to the pool
						d.pool.Put(d.events[accessed[last].T].JSON)
						// delete the map entry of the expired event json
						delete(d.events, accessed[last].T)
					}
				}
			}
		}

	}()
	return d
}

// Put stores an event's encoded JSON form for access by concurrent client
// requests. Call this with an event decoded from the database so that
// concurrent queries that match it can avoid repeated decode/allocate steps.
func (d *Encoder) Put(ev *event.T) (j []byte, err error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	buf := bytes.NewBuffer(d.pool.Get().([]byte))
	ev.ToObject().Buffer(buf)
	j = buf.Bytes()
	d.events[ev.ID] = &Event{JSON: j, lastAccessed: timestamp.Now()}
	return
}

// Get retrieves the encoded JSON for a given event if it is cached. If it
// returns a nil slice, the caller should fetch the event binary from the
// database and use Put to store it for the next call.
func (d *Encoder) Get(evId eventid.T) (b []byte) {
	d.mx.Lock()
	defer d.mx.Unlock()
	var ok bool
	var e *Event
	if e, ok = d.events[evId]; !ok {
		return
	}
	b = e.JSON
	d.events[evId].lastAccessed = timestamp.Now()
	return
}