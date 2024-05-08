package app

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/Hubmakerlabs/replicatr/pkg/nostr/context"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/event"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/eventstore"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/kind"
)

// Import a collection of JSON events from stdin or from one or more files, line
// structured JSON.
func (rl *Relay) Import(db eventstore.Store, files []string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	log.D.Ln("running import subcommand on these files:", files)
	var err error
	var fh *os.File
	buf := make([]byte, rl.MaxMessageSize)
	for i := range files {
		log.I.Ln("importing from file", files[i])
		if fh, err = os.OpenFile(files[i], os.O_RDONLY, 0755); chk.D(err) {
			continue
		}
		scanner := bufio.NewScanner(fh)
		scanner.Buffer(buf, 500000000)
		var counter int
		for scanner.Scan() {
			select {
			case <-rl.Ctx.Done():
				return
			default:
			}
			b := scanner.Bytes()
			counter++
			// log.I.Ln(counter, string(b))
			ev := &event.T{}
			if err = json.Unmarshal(b, ev); chk.E(err) {
				continue
			}
			// log.I.Ln(counter, ev.ToObject().String())
			if ev.Kind == kind.Deletion {
				// this always returns "blocked: " whenever it returns an error
				// if err = rl.handleDeleteRequest(context.Bg(), ev); chk.E(err) {
				// 	continue
				// }
			} else {
				// log.D.Ln("adding event", counter, ev.ID)
				// this will also always return a prefixed reason
				err = db.SaveEvent(context.Bg(), ev)
			}

			// evb := ev.ToCanonical().Bytes()
			// hash := sha256.Sum256(evb)
			// id := hex.Enc(hash[:])
			// if id != ev.ID.String() {
			// 	log.D.F("id mismatch got %s, expected %s", id, ev.ID.String())
			// 	continue
			// }
			// log.D.Ln("ID was valid")
			// // check signature
			// var ok bool
			// if ok, err = ev.CheckSignature(); chk.E(err) {
			// 	log.E.F("error: failed to verify signature: %v", err)
			// 	continue
			// } else if !ok {
			// 	log.E.Ln("invalid: signature is invalid")
			// 	return
			// }
			// log.D.Ln("signature was valid")
			// if ev.Kind == kind.Deletion {
			// 	// this always returns "blocked: " whenever it returns an error
			// 	// if err = rl.handleDeleteRequest(context.Bg(), ev); chk.E(err) {
			// 	// 	continue
			// 	// }
			// } else {
			// 	log.D.Ln("adding event", ev.ID)
			// 	// this will also always return a prefixed reason
			// 	err = db.SaveEvent(context.Bg(), ev)
			// }
			// chk.E(err)
		}
		chk.D(fh.Close())
	}
}
