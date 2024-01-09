package OK

import (
	"fmt"
	"sort"
	"strings"

	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/enveloper"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/envelopes/labels"
	"github.com/Hubmakerlabs/replicatr/pkg/nostr/eventid"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/array"
	"github.com/Hubmakerlabs/replicatr/pkg/wire/text"
)

var log = log2.GetStd()

type Reason string

const (
	PoW         Reason = "pow"
	Duplicate   Reason = "duplicate"
	Blocked     Reason = "blocked"
	RateLimited Reason = "rate-limited"
	Invalid     Reason = "invalid"
	Error       Reason = "error"
)

var _ enveloper.I = (*Envelope)(nil)

// Envelope is a relay message sent in response to an EventEnvelope to
// indicate acceptance (OK is true), rejection and provide a human readable
// Reason for clients to display to users, with the first word being a machine
// readable reason type, as listed in the RejectReason* constants above,
// followed by ": " and a human readable message.
type Envelope struct {
	EventID eventid.EventID
	OK      bool
	Reason  string
}

func NewOKEnvelope(eventID eventid.EventID, ok bool, reason string) (o *Envelope,
	e error) {
	var ei eventid.EventID
	if ei, e = eventid.NewEventID(string(eventID)); log.Fail(e) {
		return
	}
	o = &Envelope{
		EventID: ei,
		OK:      ok,
		Reason:  reason,
	}
	return
}

func (env *Envelope) Label() (l string) { return labels.OK }

func (env *Envelope) ToArray() (a array.T) {
	return array.T{labels.OK, env.EventID, env.OK, env.Reason}
}

func (env *Envelope) String() (s string) { return env.ToArray().String() }

func (env *Envelope) Bytes() (s []byte) { return env.ToArray().Bytes() }

// MarshalJSON returns the JSON encoded form of the envelope.
func (env *Envelope) MarshalJSON() ([]byte, error) { return env.Bytes(), nil }

const (
	Btrue     = "true"
	BtrueLen  = len(Btrue)
	Bfalse    = "false"
	BfalseLen = len(Bfalse)
)

// Unmarshal the envelope.
func (env *Envelope) Unmarshal(buf *text.Buffer) (e error) {
	log.D.Ln("ok envelope unmarshal", string(buf.Buf))
	if env == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	// Next, find the comma after the label
	if e = buf.ScanThrough(','); e != nil {
		return
	}
	// next comes an event ID
	if e = buf.ScanThrough('"'); e != nil {
		return
	}
	var eventID []byte
	if eventID, e = buf.ReadUntil('"'); log.Fail(e) {
		return fmt.Errorf("did not find event ID value in ok envelope")
	}
	// check event is a valid length
	if len(eventID) != 64 {
		return fmt.Errorf("event ID in ok envelope invalid length: %d '%s'",
			len(eventID)-1, string(eventID))
	}
	eventID = eventID[:]
	// check that it's actually hexadecimal
	const hexChars = "0123456789abcdefABCDEF"
	tmp := make([]byte, 64)
	copy(tmp, eventID)
	// this sort is backwards because invalid characters are more likely after
	// the set of hex numbers than before, and the error will be found sooner
	// and shorten the iteration below.
	sort.Slice(tmp, func(i, j int) bool { return tmp[i] > tmp[j] })
next:
	for j := range tmp {
		inSet := false
		for i := range hexChars {
			if hexChars[i] == tmp[j] {
				inSet = true
				continue next
			}
		}
		// if a character in tmp didn't match by the end of hexChars we found an invalid character.
		if !inSet {
			return fmt.Errorf("found non-hex character in event ID: '%s'",
				string(eventID))
		}
	}
	env.EventID = eventid.EventID(eventID)
	// next another comma
	if e = buf.ScanThrough(','); e != nil {
		return
	}
	// next comes a boolean value
	var isOK []byte
	if isOK, e = buf.ReadUntil(','); log.Fail(e) {
		return fmt.Errorf("did not find OK value in ok envelope")
	}
	isOK = isOK[:]
	// determine the value encoded
	l := len(isOK)
	var isBool bool
maybeOK:
	switch {
	case l == BtrueLen:
		for i := range isOK {
			if isOK[i] != Btrue[i] {
				break maybeOK
			}
		}
		env.OK = true
		isBool = true
	case l == BfalseLen:
		for i := range isOK {
			if isOK[i] != Bfalse[i] {
				break maybeOK
			}
		}
		isBool = true
	}
	if !isBool {
		return fmt.Errorf("unexpected string in ok envelope OK field '%s'",
			string(isOK))
	}
	// Next must be a string, which can be empty, but must be at minimum a pair
	// of quotes.
	if e = buf.ScanThrough('"'); e != nil {
		return
	}
	var reason []byte
	if reason, e = buf.ReadUntil('"'); log.Fail(e) {
		return fmt.Errorf("did not find reason value in ok envelope")
	}
	// Scan for the proper envelope ending.
	if e = buf.ScanThrough(']'); e != nil {
		log.D.Ln("envelope unterminated but all fields found")
	}
	env.Reason = string(text.UnescapeByteString(reason))
	return
}

// Message takes a string message that is to be sent in an `OK` or `CLOSED`
// command and prefixes it with "<prefix>: " if it doesn't already have an
// acceptable prefix.
func Message(reason Reason, prefix string) string {
	if idx := strings.Index(string(reason),
		": "); idx == -1 || strings.IndexByte(string(reason[0:idx]),
		' ') != -1 {
		return prefix + ": " + string(reason)
	}
	return string(reason)
}
