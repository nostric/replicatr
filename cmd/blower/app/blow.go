package app

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/Hubmakerlabs/replicatr/app"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/client"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/context"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/eventenvelope"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/event"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/hex"
	"github.com/Hubmakerlabs/replicatr/pkg/units"
	"github.com/minio/sha256-simd"
)

func Blower(args *Config) int {
	c := context.Bg()
	var err error
	var upRelay *client.T
	if upRelay, err = client.Connect(c, args.UploadRelay); chk.E(err) {
		return 1
	}
	log.I.Ln("connected to upload relay", args.UploadRelay)
	// var upAuthed bool
	var fh *os.File
	if fh, err = os.OpenFile(args.SourceFile, os.O_RDONLY, 0755); chk.D(err) {
		return 1
	}
	buf := make([]byte, app.MaxMessageSize)
	scanner := bufio.NewScanner(fh)
	scanner.Buffer(buf, 500000000)
	var counter, position int
	var start int64
	for scanner.Scan() {
		counter++
		b := scanner.Bytes()
		position += len(b)
		if counter <= args.Skip {
			continue
		}
		// if len(b) > app.MaxMessageSize {
		// 	log.E.Ln("message too long", string(b))
		// 	continue
		// }
		if start == 0 {
			start = time.Now().Unix()
		}
		ev := &event.T{}
		if err = json.Unmarshal(b, ev); chk.E(err) {
			continue
		}
		// if string(b) != ev.ToObject().String() {
		// 	log.W.Ln("mismatch between original and encoded/decoded")
		// continue
		// }
		can := ev.ToCanonical().Bytes()
		id := sha256.Sum256(can)
		idh := hex.Enc(id[:])
		if idh != string(ev.ID) {
			log.W.Ln("mismatch between original and encoded/decoded", hex.Enc(id[:]), string(ev.ID))
			continue
		}
		log.I.F("%d : size: %6d bytes, position: %0.6f Gb", // \n%s\n%s",
			counter, len(b), float64(position)/float64(units.Gb),
		)
		for err = <-upRelay.Write(eventenvelope.FromRawJSON("", b)); err != nil; {
			if strings.Contains(err.Error(), "failed to flush writer") {
				return 0
			}
			if strings.Contains(err.Error(), "connection closed") {
				upRelay.Close()
				upRelay.Connection.Conn.Close()
				if upRelay, err = client.Connect(c,
					args.UploadRelay); chk.E(err) {
					return 1
				}
			}
			// todo: get authing working properly
			// if !upAuthed {
			// 	log.I.Ln("authing")
			// 	// this can fail once
			// 	select {
			// 	case <-upRelay.AuthRequired:
			// 		log.T.Ln("authing to up relay")
			// 		if err = upRelay.Auth(c,
			// 			func(evt *event.T) error {
			// 				return evt.Sign(args.SeckeyHex)
			// 			}); chk.D(err) {
			// 			return 1
			// 		}
			// 		upAuthed = true
			// 		if err = upRelay.Publish(uc, ev); chk.D(err) {
			// 			return 1
			// 		}
			// 	case <-time.After(5 * time.Second):
			// 		log.E.Ln("timed out waiting to auth")
			// 		return 1
			// 	}
			// 	log.I.Ln("authed")
			// 	return 0
			// }
			// if err = upRelay.Publish(uc, ev); chk.D(err) {
			// 	return 1
			// }
		}
		// time.Sleep(time.Second)
	}
	return 0
}
