package sentinel

import (
	"fmt"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/labels"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/text"
)

var log = log2.GetStd()

// Identify takes a byte slice and scans it as a nostr Envelope array, and
// returns the label type and a text.Buffer that is ready for the Read function
// to generate the appropriate structure.
func Identify(b []byte) (match labels.T, buf *text.Buffer, e error) {
	// The bytes must be valid JSON but we can't assume they are free of
	// whitespace... So we will use some tools.
	buf = text.NewBuffer(b)
	// First there must be an opening bracket.
	if e = buf.ScanThrough('['); e != nil {
		return
	}
	// Then a quote.
	if e = buf.ScanThrough('"'); e != nil {
		return
	}
	var candidate []byte
	if candidate, e = buf.ReadUntil('"'); e != nil {
		return
	}
	// log.D.F("label: '%s' %v", string(candidate), List)
	var differs bool
matched:
	for i := range labels.List {
		differs = false
		if len(candidate) == len(labels.List[i]) {
			for j := range candidate {
				if candidate[j] != labels.List[i][j] {
					differs = true
					break
				}
			}
			if !differs {
				// there can only be one!
				match = i
				break matched
			}
		}
	}
	// If there was no match we still have zero.
	if match == labels.LNil {
		// no match
		e = fmt.Errorf("label '%s' not recognised as envelope label",
			string(candidate))
		// label is the string that was found in the first element of the JSON
		// array.
		return
	}
	return
}
