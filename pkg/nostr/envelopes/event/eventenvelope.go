package event

import (
	"encoding/json"
	"fmt"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/enveloper"
	l "github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/labels"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/event"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/subscriptionid"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/array"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/text"
)

var log = log2.GetStd()

var _ enveloper.I = (*Envelope)(nil)

// Envelope is the wrapper expected by a relay around an event.
type Envelope struct {

	// The SubscriptionID field is optional, and may at most contain 64 characters,
	// sufficient for encoding a 256 bit hash as hex.
	SubscriptionID subscriptionid.T

	// The Event is here a pointer because it should not be copied unnecessarily.
	Event *event.T
}

func (env *Envelope) UnmarshalJSON(bytes []byte) error {
	// TODO implement me
	panic("implement me")
}

// NewEventEnvelope builds an Envelope from a provided T
// string and pointer to an T, and returns either the Envelope or an
// error if the Subscription ID is invalid or the T is nil.
func NewEventEnvelope(si string, ev *event.T) (ee *Envelope, e error) {
	var sid subscriptionid.T
	if sid, e = subscriptionid.New(si); log.Fail(e) {
		return
	}
	if ev == nil {
		e = fmt.Errorf("cannot make event envelope with nil event")
		return
	}
	return &Envelope{SubscriptionID: sid, Event: ev}, nil
}

func (env *Envelope) Label() string { return l.EVENT }

func (env *Envelope) ToArray() (a array.T) {
	a = make(array.T, 0, 3)
	a = append(a, l.EVENT)
	if env.SubscriptionID.IsValid() {
		a = append(a, env.SubscriptionID)
	}
	a = append(a, env.Event.ToObject())
	return
}

func (env *Envelope) String() (s string) { return env.ToArray().String() }

func (env *Envelope) Bytes() (s []byte) { return env.ToArray().Bytes() }

func (env *Envelope) MarshalJSON() ([]byte, error) { return env.Bytes(), nil }

// Unmarshal the envelope.
func (env *Envelope) Unmarshal(buf *text.Buffer) (e error) {
	if env == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	// Next, find the comma after the label (note we aren't checking that only
	// whitespace intervenes because laziness, usually this is the very next
	// character).
	if e = buf.ScanUntil(','); log.D.Chk(e) {
		return
	}
	// Next character we find will be open quotes for the subscription ID, or
	// the open brace of the embedded event.
	var matched byte
	if matched, e = buf.ScanForOneOf(false, '"', '{'); log.D.Chk(e) {
		return
	}
	if matched == '"' {
		// Advance the cursor to consume the quote character.
		buf.Pos++
		var sid []byte
		// Read the string.
		if sid, e = buf.ReadUntil('"'); log.D.Chk(e) {
			return fmt.Errorf("unterminated quotes in JSON, probably truncated read")
		}
		env.SubscriptionID = subscriptionid.T(sid[:])
		// Next, find the comma after the subscription ID (note we aren't checking
		// that only whitespace intervenes because laziness, usually this is the
		// very next character).
		if e = buf.ScanUntil(','); log.D.Chk(e) {
			return fmt.Errorf("event not found in event envelope")
		}
	}
	// find the opening brace of the event object, usually this is the very next
	// character, we aren't checking for valid whitespace because laziness.
	if e = buf.ScanUntil('{'); log.D.Chk(e) {
		return fmt.Errorf("event not found in event envelope")
	}
	// now we should have an event object next. It has no embedded object so it
	// should end with a close brace. This slice will be wrapped in braces and
	// contain paired brackets, braces and quotes.
	var eventObj []byte
	if eventObj, e = buf.ReadEnclosed(); log.Fail(e) {
		return
	}
	// allocate an event to unmarshal into
	env.Event = &event.T{}
	if e = json.Unmarshal(eventObj, env.Event); log.Fail(e) {
		log.D.S(string(eventObj))
		return
	}
	// technically we maybe should read ahead further to make sure the JSON
	// closes correctly. Not going to abort because of this.
	if e = buf.ScanUntil(']'); log.D.Chk(e) {
		return fmt.Errorf("malformed JSON, no closing bracket on array")
	}
	// whatever remains doesn't matter as the envelope has fully unmarshaled.
	return
}
