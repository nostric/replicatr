package countrequest

import (
	"encoding/json"
	"fmt"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	l "github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/labels"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/filter"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/filters"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/subscriptionid"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/array"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/text"
)

var log = log2.GetStd()

type Envelope struct {
	SubscriptionID subscriptionid.T
	filters.T
}

func (C *Envelope) Label() string { return l.COUNT }

func (C *Envelope) ToArray() array.T {
	return array.T{l.COUNT, C.SubscriptionID, C.T}
}

func (C *Envelope) String() string { return C.ToArray().String() }

func (C *Envelope) Bytes() []byte { return C.ToArray().Bytes() }

func (C *Envelope) MarshalJSON() ([]byte, error) { return C.Bytes(), nil }

func (C *Envelope) Unmarshal(buf *text.Buffer) (e error) {
	log.D.Ln("ok envelope unmarshal", string(buf.Buf))
	if C == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	// Next, find the comma after the label.
	if e = buf.ScanThrough(','); e != nil {
		return
	}
	// Next character we find will be open quotes for the subscription ID.
	if e = buf.ScanThrough('"'); e != nil {
		return
	}
	var sid []byte
	// read the string
	if sid, e = buf.ReadUntil('"'); log.Fail(e) {
		return fmt.Errorf("unterminated quotes in JSON, probably truncated read")
	}
	C.SubscriptionID = subscriptionid.T(sid)
	// find the opening brace of the first or only filter object.
	if e = buf.ScanUntil('{'); e != nil {
		return fmt.Errorf("event not found in event envelope")
	}
	// T in the count envelope are variadic, there can be more than one,
	// with subsequent items separated by a comma, so we read them in in a loop,
	// breaking when we don't find a comma after.
	for {
		var filterArray []byte
		if filterArray, e = buf.ReadEnclosed(); log.Fail(e) {
			return
		}
		f := &filter.T{}
		if e = json.Unmarshal(filterArray, &f); log.Fail(e) {
			return
		}
		C.T = append(C.T, f)
		cur := buf.Pos
		// Next, find the comma after filter.
		if e = buf.ScanThrough(','); e != nil {
			// we didn't find one, so break the loop.
			buf.Pos = cur
			break
		}
	}
	// If we found at least one filter, there is no error, the io.EOF is
	// expected at any point after at least one filter.
	if len(C.T) > 0 {
		e = nil
	}
	// // Technically we maybe should read ahead further to make sure the JSON
	// // closes correctly. Not going to abort because of this.
	// //
	// TODO: this is a waste of time really, the rest of the buffer will be
	//  discarded anyway as no more content is expected
	// if e = buf.ScanUntil(']'); e != nil {
	// 	return fmt.Errorf("malformed JSON, no closing bracket on array")
	// }
	// whatever remains doesn't matter as the envelope has fully unmarshaled.
	return
}
